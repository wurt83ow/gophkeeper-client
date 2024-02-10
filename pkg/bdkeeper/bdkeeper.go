/*
Package bdkeeper provides functionalities to interact with a SQLite database
and perform operations like user management, data manipulation, and synchronization.

Usage:

	import "github.com/example.com/bdkeeper"

	keeper := bdkeeper.NewKeeper()

	// Example: Check if a user exists
	exists, err := keeper.UserExists(context.Background(), "username")
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println("User exists:", exists)

	// Example: Add a new user
	err := keeper.AddUser(context.Background(), "username", "hashedPassword")
	if err != nil {
	    log.Fatal(err)
	}

	// Other functionalities are used similarly.

Data Structures:

	Keeper - Represents a database keeper with methods to interact with the database.

	SyncQueue - Represents a synchronization queue entry with pending operations.

Example Usage:

	keeper := bdkeeper.NewKeeper()

	// Check if a user exists
	exists, err := keeper.UserExists(context.Background(), "username")
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println("User exists:", exists)

	// Add a new user
	err := keeper.AddUser(context.Background(), "username", "hashedPassword")
	if err != nil {
	    log.Fatal(err)
	}

	// Other functionalities are used similarly.

Thread Safety:

	The methods provided by Keeper are not thread-safe. It is expected that
	the user of the package ensures proper synchronization when used concurrently.
*/
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

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// embeddedMigrations contains SQL migration files.
//
//go:embed migrations/*.sql
var embeddedMigrations embed.FS

// Keeper represents a database keeper with methods to interact with the database.
type Keeper struct {
	db *sql.DB
}

// NewKeeper creates a new instance of Keeper and initializes the database.
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

	// Apply migrations
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

// UserExists checks if a user exists in the database.
func (k *Keeper) UserExists(ctx context.Context, username string) (bool, error) {
	// Query to check the existence of a user in the database
	query := `SELECT COUNT(*) FROM Users WHERE username = ?;`

	// Execute the query
	row := k.db.QueryRowContext(ctx, query, username)

	// Get the result
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	// If the count is greater than 0, the user exists
	return count > 0, nil
}

// AddUser adds a new user to the database.
func (k *Keeper) AddUser(ctx context.Context, username string, hashedPassword string) error {
	// Query to add a new user to the database
	query := `INSERT INTO Users (username, password) VALUES (?, ?);`

	// Execute the query
	_, err := k.db.ExecContext(ctx, query, username, hashedPassword)
	return err
}

// IsEmpty checks if the database is empty.
func (k *Keeper) IsEmpty(ctx context.Context) (bool, error) {
	// Query to get the count of entries in all tables
	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';`

	// Execute the query
	row := k.db.QueryRowContext(ctx, query)

	// Get the result
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	// If the count is 0, the database is empty
	return count == 0, nil
}

// CreateSyncEntry creates a synchronization queue entry in the database.
func (k *Keeper) CreateSyncEntry(ctx context.Context, operation string, table string, user_id int, entry_id string, data map[string]string) error {
	// Convert data to JSON
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set initial status as "Pending"
	status := "Pending"

	// Add entry to SyncQueue table
	_, err = k.db.ExecContext(ctx, "INSERT INTO SyncQueue (operation, table_name, user_id, entry_id, data, status) VALUES (?, ?, ?, ?, ?, ?)",
		operation, table, user_id, entry_id, dataJson, status)
	return err
}

// GetPassword retrieves the hashed password of a user from the database.
func (k *Keeper) GetPassword(ctx context.Context, username string) (string, error) {
	// Query to get the hashed password of a user from the database
	query := `SELECT password FROM Users WHERE username = ?;`

	// Execute the query
	row := k.db.QueryRowContext(ctx, query, username)

	// Get the result
	var password string
	err := row.Scan(&password)
	if err != nil {
		return "", err
	}

	// Return the hashed password
	return password, nil
}

// CompareHashAndPassword compares a hashed password with a plaintext password.
func (k *Keeper) CompareHashAndPassword(hashedPassword, password string) bool {
	// Compare the hashed password with the hash of the entered password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetUserID retrieves the user ID of a user from the database.
func (k *Keeper) GetUserID(ctx context.Context, username string) (int, error) {
	// Query to get the user ID from the database
	query := `SELECT id FROM Users WHERE username = ?;`

	// Execute the query
	row := k.db.QueryRowContext(ctx, query, username)

	// Get the result
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	// Return the user ID
	return id, nil
}

// AddData adds data to the specified database table.
// Keys and values from the provided data map will be inserted into the specified table, along with the given user_id and entry_id.
func (k *Keeper) AddData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
	// Create lists of keys and values considering user_id and entry_id
	keys := make([]string, 0, len(data)+2)        // +2 for user_id and entry_id
	values := make([]interface{}, 0, len(data)+2) // +2 for user_id and entry_id

	// Add user_id and entry_id to the beginning of keys and values lists
	keys = append(keys, "user_id", "id")
	values = append(values, user_id, entry_id)

	// Add other keys and values from the provided data map
	for key, value := range data {
		if key == "deleted" {
			continue
		}
		keys = append(keys, key)
		values = append(values, value)
	}

	// Prepare the SQL query and execute it
	stmt, err := k.db.Prepare(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(keys, ","), strings.Repeat("?,", len(keys)-1)+"?"))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)

	return err
}

// UpdateData updates data in the specified database table.
// Values from the provided data map will be used to update the record with the specified user_id and entry_id in the specified table.
func (k *Keeper) UpdateData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
	// Create lists of keys and values to update data
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for key, value := range data {
		keys = append(keys, key+" = ?")
		values = append(values, value)
	}
	// Append user_id and entry_id to the end of the lists
	values = append(values, user_id, entry_id)

	// Prepare the SQL query and execute it
	stmt, err := k.db.Prepare(fmt.Sprintf("UPDATE %s SET %s WHERE user_id = ? AND id = ?", table, strings.Join(keys, ",")))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, values...)
	return err
}

// DeleteData deletes a record from the specified table in the database.
// It checks for the existence of the specified user_id and entry_id, and then deletes the record if found.
func (k *Keeper) DeleteData(ctx context.Context, table string, user_id int, entry_id string) error {
	// Check user_id and table
	if user_id == 0 || table == "" {
		return errors.New("user_id and table must be specified")
	}

	// Check entry_id
	if entry_id == "" {
		return errors.New("id must be specified")
	}

	// Prepare the query to check the existence of the record
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

// GetData retrieves data for a specific entry from the specified table in the database.
// It returns a map containing column names as keys and corresponding values for the entry.
func (k *Keeper) GetData(ctx context.Context, table string, user_id int, entry_id string) (map[string]string, error) {
	// Get all columns of the table
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
		// Exclude unnecessary columns
		if col.Name != "id" && col.Name != "deleted" && col.Name != "user_id" && col.Name != "updated_at" {
			cols = append(cols, col.Name)
		}
	}

	// Query the specific entry
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

// GetAllData retrieves all data from the specified table in the database for the given user_id.
// It returns a slice of maps, with each map containing column names as keys and corresponding values for each row.
func (k *Keeper) GetAllData(ctx context.Context, table string, user_id int, columns ...string) ([]map[string]string, error) {
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

// ClearData deletes all records associated with the specified user_id from the specified table in the database.
func (k *Keeper) ClearData(ctx context.Context, table string, userID int) error {
	stmt, err := k.db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", table))
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, userID)
	return err
}

// GetPendingSyncEntries returns all entries from the sync table with status "Pending".
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

// UpdateSyncEntryStatus updates the status of an entry in the sync table.
func (k *Keeper) UpdateSyncEntryStatus(ctx context.Context, id int, status string) error {
	_, err := k.db.ExecContext(ctx, "UPDATE SyncQueue SET status = ? WHERE id = ?", status, id)
	return err
}
