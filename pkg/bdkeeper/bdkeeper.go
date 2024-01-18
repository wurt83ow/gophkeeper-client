package bdkeeper

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Keeper struct {
	db *sql.DB
}

func New() *Keeper {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	return &Keeper{
		db: db,
	}
}
func (k *Keeper) AddData(user_id int, table string, data map[string]string) {
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for key, value := range data {
		keys = append(keys, key)
		values = append(values, value)
	}
	stmt, _ := k.db.Prepare(fmt.Sprintf("INSERT INTO %s(%s) values(%s)", table, strings.Join(keys, ","), strings.Repeat("?,", len(keys)-1)+"?"))
	stmt.Exec(values...)
}

func (k *Keeper) UpdateData(user_id int, table string, data map[string]string) {
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for key, value := range data {
		keys = append(keys, key+" = ?")
		values = append(values, value)
	}
	values = append(values, user_id)
	stmt, _ := k.db.Prepare(fmt.Sprintf("UPDATE %s SET %s WHERE user_id = ?", table, strings.Join(keys, ",")))
	stmt.Exec(values...)
}

func (k *Keeper) DeleteData(user_id int, table string) {
	stmt, _ := k.db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", table))
	stmt.Exec(user_id)
}

func (k *Keeper) GetData(user_id int, table string, columns ...string) map[string]string {
	row := k.db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ?", strings.Join(columns, ","), table), user_id)
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(sql.RawBytes)
	}
	row.Scan(values...)
	data := make(map[string]string)
	for i, column := range columns {
		data[column] = string(*values[i].(*sql.RawBytes))
	}
	return data
}

func (k *Keeper) GetAllData(table string, columns ...string) []map[string]string {
	rows, _ := k.db.Query(fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), table))
	defer rows.Close()
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(sql.RawBytes)
	}
	var data []map[string]string
	for rows.Next() {
		rows.Scan(values...)
		row := make(map[string]string)
		for i, column := range columns {
			row[column] = string(*values[i].(*sql.RawBytes))
		}
		data = append(data, row)
	}
	return data
}

func (k *Keeper) ClearData(table string) {
	stmt, _ := k.db.Prepare(fmt.Sprintf("DELETE FROM %s", table))
	stmt.Exec()
}
