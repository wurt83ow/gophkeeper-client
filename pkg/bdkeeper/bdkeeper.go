package bdkeeper

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Keeper struct {
	db *sql.DB
}
type Data interface {
	Fields() ([]string, []interface{})
}

func NewKeeper() *Keeper {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}

	k := &Keeper{
		db: db,
	}

	// Create the migrations table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY)")
	if err != nil {
		panic(err)
	}

	files, err := fs.ReadDir(embeddedMigrations, "migrations")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			// Check if the migration has already been applied
			var name string
			err = db.QueryRow("SELECT name FROM migrations WHERE name = ?", file.Name()).Scan(&name)
			if err == sql.ErrNoRows {
				// The migration has not been applied yet
				f, err := embeddedMigrations.Open("migrations/" + file.Name())
				if err != nil {
					panic(err)
				}
				defer f.Close()

				bytes, err := io.ReadAll(f)
				if err != nil {
					panic(err)
				}

				upAndDown := strings.Split(string(bytes), "-- +goose Down")
				upStatements := strings.Split(upAndDown[0], ";")

				// Run the "up" statements.
				for _, stmt := range upStatements {
					if _, err := db.Exec(stmt); err != nil {
						log.Fatalf("Failed to execute migration %s: %v", file.Name(), err)
					}
				}

				// Record the migration as having been applied
				_, err = db.Exec("INSERT INTO migrations (name) VALUES (?)", file.Name())
				if err != nil {
					panic(err)
				}
			} else if err != nil {
				// An error occurred checking if the migration has been applied
				panic(err)
			}
		}
	}

	return k
}

func (k *Keeper) UserExists(ctx context.Context, username string) (bool, error) {
	// Запрос для проверки наличия пользователя в базе данных
	query := `SELECT COUNT(*) FROM Users WHERE username = ?;`

	// Выполнение запроса
	row := k.db.QueryRowContext(ctx, query, username)

	// Получение результата
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	// Если количество записей больше 0, значит пользователь существует
	return count > 0, nil
}

func (k *Keeper) AddUser(ctx context.Context, username string, hashedPassword string) error {
	// Запрос для добавления нового пользователя в базу данных
	query := `INSERT INTO Users (username, password) VALUES (?, ?);`

	// Выполнение запроса
	_, err := k.db.ExecContext(ctx, query, username, hashedPassword)
	return err
}

func (k *Keeper) IsEmpty(ctx context.Context) (bool, error) {
	// Запрос для получения количества записей во всех таблицах
	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';`

	// Выполнение запроса
	row := k.db.QueryRowContext(ctx, query)

	// Получение результата
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	// Если количество записей равно 0, значит база данных пуста
	return count == 0, nil
}

func (k *Keeper) MarkForSync(ctx context.Context, user_id int, table string, data Data) error {
	// Получить данные из интерфейса Data
	keys, values := data.Fields()

	// Создать map для преобразования в JSON
	dataMap := make(map[string]interface{})
	for i, key := range keys {
		dataMap[key] = values[i]
	}

	// Преобразовать данные в JSON
	dataJson, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}

	// Добавить запись в таблицу SyncQueue
	_, err = k.db.ExecContext(ctx, "INSERT INTO SyncQueue (user_id, table_name, data) VALUES (?, ?, ?)", user_id, table, dataJson)
	return err
}

func (k *Keeper) GetPassword(ctx context.Context, username string) (string, error) {
	// Запрос для получения хешированного пароля пользователя из базы данных
	query := `SELECT password FROM Users WHERE username = ?;`

	// Выполнение запроса
	row := k.db.QueryRowContext(ctx, query, username)

	// Получение результата
	var password string
	err := row.Scan(&password)
	if err != nil {
		return "", err
	}

	// Возвращаем хешированный пароль
	return password, nil
}

func (k *Keeper) CompareHashAndPassword(hashedPassword, password string) bool {
	// Сравнение хешированного пароля с хешем введенного пароля
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (k *Keeper) GetUserID(ctx context.Context, username string) (int, error) {
	// Запрос для получения идентификатора пользователя из базы данных
	query := `SELECT id FROM Users WHERE username = ?;`

	// Выполнение запроса
	row := k.db.QueryRowContext(ctx, query, username)

	// Получение результата
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	// Возвращаем идентификатор пользователя
	return id, nil
}

func (k *Keeper) AddData(ctx context.Context, user_id int, table string, id string, data Data) error {
	keys, values := data.Fields()

	// Добавьте id и user_id в начало списков ключей и значений
	keys = append([]string{"id", "user_id"}, keys...)
	values = append([]interface{}{id, user_id}, values...)

	stmt, err := k.db.Prepare(fmt.Sprintf("INSERT INTO %s(%s) values(%s)", table, strings.Join(keys, ","), strings.Repeat("?,", len(keys)-1)+"?"))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)
	return err
}

func (k *Keeper) UpdateData(ctx context.Context, user_id int, table string, id string, data Data) error {
	keys, values := data.Fields()
	setClauses := make([]string, len(keys))
	for i, key := range keys {
		setClauses[i] = key + " = ?"
	}
	values = append(values, user_id, id)
	stmt, err := k.db.Prepare(fmt.Sprintf("UPDATE %s SET %s WHERE user_id = ? AND id = ?", table, strings.Join(setClauses, ",")))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)
	return err
}

func (k *Keeper) DeleteData(ctx context.Context, user_id int, table string, id string) error {
	// Check user_id and table
	if user_id == 0 || table == "" {
		return errors.New("user_id and table must be specified")
	}

	// Check id
	if id == "" {
		return errors.New("id must be specified")
	}

	// Prepare the query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND id = ?", table)
	args := []interface{}{user_id, id}

	// Execute the query
	row := k.db.QueryRowContext(ctx, query, args...)
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
	_, err = k.db.ExecContext(ctx, query, args...)
	return err
}

func (k *Keeper) GetData(ctx context.Context, user_id int, table string, id string, data Data) error {
	// Получаем все колонки таблицы
	columns, err := k.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	var cols []string
	for columns.Next() {
		var col struct {
			Cid        int
			Name       string
			Type       string
			NotNull    bool
			Dflt_value *string
			Pk         int
		}
		err := columns.Scan(&col.Cid, &col.Name, &col.Type, &col.NotNull, &col.Dflt_value, &col.Pk)
		if err != nil {
			return fmt.Errorf("failed to scan column: %w", err)
		}
		// Исключаем ненужные столбцы
		if col.Name != "id" && col.Name != "deleted" && col.Name != "user_id" && col.Name != "updated_at" {
			cols = append(cols, col.Name)
		}
	}

	row := k.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND deleted = false", strings.Join(cols, ","), table), id)
	values := make([]interface{}, len(cols))
	for i := range values {
		var value string
		values[i] = &value
	}
	err = row.Scan(values...)
	if err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Заполняем данные в структуре Data
	fields, _ := data.Fields()
	for i, column := range cols {
		for j, field := range fields {
			if field == column {
				switch v := values[i].(type) {
				case *string:
					*values[j].(*string) = *v
				case *int:
					*values[j].(*int) = *v
				}
			}
		}
	}

	return nil
}

func (k *Keeper) GetAllData(ctx context.Context, table string, columns ...string) ([]map[string]string, error) {

	rows, err := k.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), table))
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

func (k *Keeper) ClearData(ctx context.Context, userID int, table string) error {
	stmt, err := k.db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", table))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, userID)
	return err
}
