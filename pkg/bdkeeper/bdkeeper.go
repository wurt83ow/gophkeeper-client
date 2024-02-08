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
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Keeper struct {
	db *sql.DB
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

func (k *Keeper) CreateSyncEntry(ctx context.Context, operation string, table string,
	user_id int, entry_id string, data map[string]string) error {
	// Преобразовать данные в JSON
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Устанавливаем начальный статус "Pending"
	status := "Pending"

	// Добавить запись в таблицу SyncQueue
	_, err = k.db.ExecContext(ctx, "INSERT INTO SyncQueue (operation, table_name, user_id, entry_id, data, status) VALUES (?, ?, ?, ?, ?, ?)",
		operation, table, user_id, entry_id, dataJson, status)
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

func (k *Keeper) AddData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
	keys := make([]string, 0, len(data)+2)        // +2 для user_id и entry_id
	values := make([]interface{}, 0, len(data)+2) // +2 для user_id и entry_id

	// Добавьте user_id и entry_id в начало списков ключей и значений
	keys = append(keys, "user_id", "id")
	values = append(values, user_id, entry_id)

	for key, value := range data {
		keys = append(keys, key)
		values = append(values, value)
	}
	stmt, err := k.db.Prepare(fmt.Sprintf("INSERT INTO %s(%s) values(%s)", table, strings.Join(keys, ","), strings.Repeat("?,", len(keys)-1)+"?"))
	if err != nil {
		fmt.Println("444444444444444444444444444444444444444444444444444444444", err)
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)

	return err
}

func (k *Keeper) UpdateData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for key, value := range data {
		keys = append(keys, key+" = ?")
		values = append(values, value)
	}
	values = append(values, user_id, entry_id)
	stmt, err := k.db.Prepare(fmt.Sprintf("UPDATE %s SET %s WHERE user_id = ? AND id = ?", table, strings.Join(keys, ",")))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)
	return err
}

func (k *Keeper) DeleteData(ctx context.Context, table string, user_id int, entry_id string) error {
	// Check user_id and table
	if user_id == 0 || table == "" {
		return errors.New("user_id and table must be specified")
	}

	// Check id
	if entry_id == "" {
		return errors.New("id must be specified")
	}

	// Prepare the query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND id = ?", table)
	args := []interface{}{user_id, entry_id}

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

func (k *Keeper) GetData(ctx context.Context, table string, user_id int, entry_id string) (map[string]string, error) {
	// Получаем все колонки таблицы
	columns, err := k.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
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
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		// Исключаем ненужные столбцы
		if col.Name != "id" && col.Name != "deleted" && col.Name != "user_id" && col.Name != "updated_at" {
			cols = append(cols, col.Name)
		}
	}

	row := k.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(cols, ","), table), entry_id)
	values := make([]interface{}, len(cols))
	for i := range values {
		var value string
		values[i] = &value
	}
	err = row.Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	data := make(map[string]string)
	for i, column := range cols {
		data[column] = *(values[i].(*string))
	}
	return data, nil
}

func (k *Keeper) GetAllData(ctx context.Context, table string, user_id int, columns ...string) ([]map[string]string, error) {

	fmt.Println("sfdlsjdflkjsdlkfjlskdjflkjsdlkfjs", table, user_id)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ?", strings.Join(columns, ","), table)
	rows, err := k.db.QueryContext(ctx, query, user_id)
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

func (k *Keeper) ClearData(ctx context.Context, table string, userID int) error {
	stmt, err := k.db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", table))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, userID)
	return err
}

// GetPendingSyncEntries возвращает все записи из таблицы синхронизации со статусом "Pending"
func (k *Keeper) GetPendingSyncEntries(ctx context.Context) ([]models.SyncQueue, error) {
	var entries []models.SyncQueue

	rows, err := k.db.QueryContext(ctx, "SELECT * FROM SyncQueue WHERE status = 'Pending'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry models.SyncQueue
		err = rows.Scan(&entry.ID, &entry.TableName, &entry.UserID, &entry.EntryID, &entry.Operation, &entry.Data, &entry.Status)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// UpdateSyncEntryStatus обновляет статус записи в таблице синхронизации
func (k *Keeper) UpdateSyncEntryStatus(ctx context.Context, id int, status string) error {
	_, err := k.db.ExecContext(ctx, "UPDATE SyncQueue SET status = ? WHERE id = ?", status, id)
	return err
}
