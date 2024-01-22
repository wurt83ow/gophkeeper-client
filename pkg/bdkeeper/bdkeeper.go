package bdkeeper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

type Keeper struct {
	db *sql.DB
}

func NewKeeper() *Keeper {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	// Save old output
	old := log.Writer()

	// Redirect output to ioutil.Discard
	log.SetOutput(io.Discard)
	err = goose.Up(db, "migrations")
	// Restore the old output

	log.SetOutput(old)

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	return &Keeper{
		db: db,
	}
}

func (k *Keeper) MarkForSync(user_id int, table string, data map[string]string) error {
	// Преобразовать данные в JSON
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Добавить запись в таблицу SyncQueue
	_, err = k.db.Exec("INSERT INTO SyncQueue (user_id, table_name, data) VALUES (?, ?, ?)", user_id, table, dataJson)
	return err
}

func (k *Keeper) AddData(user_id int, table string, data map[string]string) error {
	keys := make([]string, 0, len(data)+1)        // +1 для user_id
	values := make([]interface{}, 0, len(data)+1) // +1 для user_id

	// Добавьте user_id в начало списков ключей и значений
	keys = append(keys, "user_id")
	values = append(values, user_id)

	for key, value := range data {
		keys = append(keys, key)
		values = append(values, value)
	}
	stmt, err := k.db.Prepare(fmt.Sprintf("INSERT INTO %s(%s) values(%s)", table, strings.Join(keys, ","), strings.Repeat("?,", len(keys)-1)+"?"))
	if err != nil {
		return err
	}
	_, err = stmt.Exec(values...)
	return err
}

func (k *Keeper) UpdateData(user_id int, table string, data map[string]string) error {
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for key, value := range data {
		keys = append(keys, key+" = ?")
		values = append(values, value)
	}
	values = append(values, user_id)
	stmt, err := k.db.Prepare(fmt.Sprintf("UPDATE %s SET %s WHERE user_id = ?", table, strings.Join(keys, ",")))
	if err != nil {
		return err
	}
	_, err = stmt.Exec(values...)
	return err
}

func (k *Keeper) DeleteData(user_id int, table string, id string, meta_info string) error {
	// Check user_id and table
	if user_id == 0 || table == "" {
		return errors.New("user_id and table must be specified")
	}

	// Check id and meta_info
	if id == "" && meta_info == "" {
		return errors.New("id or meta_info must be specified")
	}

	// Prepare the query
	var query string
	args := []interface{}{user_id}
	if id != "" {
		query = "id = ?"
		args = append(args, id)
	}
	if meta_info != "" {
		if id != "" {
			query += " AND "
		}
		query += "meta_info LIKE ?"
		args = append(args, "%"+meta_info+"%")
	}
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND (%s)", table, query)

	// Execute the query
	row := k.db.QueryRow(query, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	// Check the number of records found
	if count > 1 {
		return errors.New("More than one record found")
	} else if count == 0 {
		return errors.New("No records found")
	}

	// Delete the record
	query = strings.Replace(query, "SELECT COUNT(*)", "DELETE", 1)
	_, err = k.db.Exec(query, args...)
	return err
}

func (k *Keeper) GetData(user_id int, table string, columns ...string) (map[string]string, error) {
	row := k.db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ?", strings.Join(columns, ","), table), user_id)
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(sql.RawBytes)
	}
	err := row.Scan(values...)
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for i, column := range columns {
		data[column] = string(*values[i].(*sql.RawBytes))
	}
	return data, nil
}

func (k *Keeper) GetAllData(table string, columns ...string) ([]map[string]string, error) {

	rows, err := k.db.Query(fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), table))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	values := make([]interface{}, len(columns))

	for i := range values {
		values[i] = new(sql.RawBytes)
	}

	var data []map[string]string
	for rows.Next() {
		err := rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]string)
		for i, column := range columns {
			row[column] = string(*values[i].(*sql.RawBytes))
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows encountered an error: %w", err)
	}

	return data, nil
}

func (k *Keeper) ClearData(table string) error {
	stmt, err := k.db.Prepare(fmt.Sprintf("DELETE FROM %s", table))
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}
