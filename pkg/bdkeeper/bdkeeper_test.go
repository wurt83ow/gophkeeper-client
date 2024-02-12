package bdkeeper_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
)

func setup(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Create Users table
	_, err = db.Exec(`CREATE TABLE Users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	)`)
	if err != nil {
		t.Fatalf("failed to create Users table: %v", err)
	}

	// Create UserCredentials table
	_, err = db.Exec(`CREATE TABLE UserCredentials (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		login TEXT NOT NULL,
		password TEXT NOT NULL,
		meta_info TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES Users(id)
	)`)
	if err != nil {
		t.Fatalf("failed to create UserCredentials table: %v", err)
	}

	// Create SyncQueue table
	_, err = db.Exec(`CREATE TABLE SyncQueue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		table_name TEXT NOT NULL,
		user_id INTEGER,  
		entry_id TEXT,
		operation TEXT NOT NULL CHECK(operation IN ('Create', 'Update', 'Delete')),
		data TEXT NOT NULL,
		status TEXT NOT NULL CHECK(status IN ('Pending', 'Progress', 'Done', 'Error'))
	)`)
	if err != nil {
		t.Fatalf("failed to create SyncQueue table: %v", err)
	}

	// Cleanup function to close the database
	cleanup := func() {
		err := db.Close()
		if err != nil {
			t.Logf("error closing database: %v", err)
		}
	}

	return db, cleanup
}

func TestIsEmpty_NonEmptyDB(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Add a table to the database to make it non-empty
	_, err := db.Exec("CREATE TABLE Test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	isEmpty, err := keeper.IsEmpty(ctx)
	if err != nil {
		t.Fatalf("IsEmpty returned error: %v", err)
	}

	if isEmpty {
		t.Error("IsEmpty should return false for a non-empty database")
	}
}

// TestUserExists_UserExists проверяет, что метод UserExists возвращает true для существующего пользователя.
func TestUserExists_UserExists(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем новый экземпляр Keeper
	keeper := bdkeeper.NewKeeper(db)

	// Добавляем пользователя для тестирования
	_, err := db.Exec("INSERT INTO Users (username, password) VALUES (?, ?)", "testuser", "hashedpassword")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем существование пользователя
	exists, err := keeper.UserExists(context.Background(), "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что пользователь существует
	if !exists {
		t.Errorf("UserExists returned false for an existing user")
	}
}

// TestAddUser_AddsUser проверяет, что метод AddUser успешно добавляет нового пользователя.
func TestAddUser_AddsUser(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем новый экземпляр Keeper
	keeper := bdkeeper.NewKeeper(db)

	// Добавляем пользователя
	err := keeper.AddUser(context.Background(), "newuser", "hashedpassword")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что пользователь был добавлен
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ?", "newuser").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Expected 1 user to be added, got %d", count)
	}
}

func TestCreateSyncEntry(t *testing.T) {
	fmt.Println("77777777777777777777777777777777")
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Create a test synchronization entry
	err := keeper.CreateSyncEntry(ctx, "Create", "UserCredentials", 1, "entry_id", map[string]string{"key": "value"})
	if err != nil {
		fmt.Println("77777777777777777777777777777777", err)
		t.Fatalf("CreateSyncEntry returned error: %v", err)
	}

	// Check if the entry exists in the database
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM SyncQueue WHERE operation = ? AND table_name = ? AND user_id = ? AND entry_id = ?", "Create", "UserCredentials", 1, "entry_id").Scan(&count)
	if err != nil {
		fmt.Println("888888888888888888888888888888888888888888888888", err)
		t.Fatalf("failed to query SyncQueue: %v", err)
	}
	fmt.Println("2222222222222222222222222222222222222", count)
	if count != 1 {
		t.Error("CreateSyncEntry failed to add synchronization entry to the database")
	}
}

func TestGetPassword(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Add a test user
	addUser(db, "testuser", "password123")

	// Retrieve the password of the test user
	password, err := keeper.GetPassword(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetPassword returned error: %v", err)
	}

	if password != "password123" {
		t.Errorf("GetPassword returned incorrect password: got %s, want password123", password)
	}
}

func TestGetUserID(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Add a test user
	addUser(db, "testuser", "password123")

	// Retrieve the ID of the test user
	id, err := keeper.GetUserID(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetUserID returned error: %v", err)
	}

	if id != 1 {
		t.Errorf("GetUserID returned incorrect user ID: got %d, want 1", id)
	}
}

func addUser(db *sql.DB, username, password string) {
	_, err := db.Exec("INSERT INTO Users (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		log.Fatalf("failed to add user: %v", err)
	}
}

func TestGetAllData_UserCredentials(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Insert test data into the UserCredentials table
	_, err := db.Exec(`INSERT INTO UserCredentials (id, user_id, login, password) VALUES ('1', 1, 'user1', 'password1'), ('2', 1, 'user2', 'password2')`)
	assert.NoError(t, err)

	// Retrieve all data from the UserCredentials table for user_id 1
	data, err := keeper.GetAllData(ctx, "UserCredentials", 1, "id", "login", "password")
	assert.NoError(t, err)

	expected := []map[string]string{
		{"id": "1", "login": "user1", "password": "password1"},
		{"id": "2", "login": "user2", "password": "password2"},
	}

	assert.Equal(t, expected, data)
}

func TestClearData_UserCredentials(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Insert test data into the UserCredentials table
	_, err := db.Exec(`INSERT INTO UserCredentials (id, user_id, login, password) VALUES ('1', 1, 'user1', 'password1'), ('2', 1, 'user2', 'password2')`)
	assert.NoError(t, err)

	// Clear data for user_id 1 from the UserCredentials table
	err = keeper.ClearData(ctx, "UserCredentials", 1)
	assert.NoError(t, err)

	// Verify that no data exists for user_id 1 in the UserCredentials table
	data, err := keeper.GetAllData(context.Background(), "UserCredentials", 1, "id", "login", "password")
	assert.NoError(t, err)
	assert.Empty(t, data)
}

func TestGetPendingSyncEntries(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Insert test data into the SyncQueue table
	_, err := db.Exec(`INSERT INTO SyncQueue (operation, table_name, user_id, entry_id, data, status) VALUES
		('Create', 'table1', 1, 'entry1', '{"key1": "value1"}', 'Pending'),
		('Update', 'table2', 2, 'entry2', '{"key2": "value2"}', 'Pending')`)

	assert.NoError(t, err)

	// Retrieve pending sync entries
	entries, err := keeper.GetPendingSyncEntries(ctx)
	assert.NoError(t, err)

	expected := []models.SyncQueue{
		{ID: 1, Operation: "Create", TableName: "table1", UserID: 1, EntryID: "entry1", Data: `{"key1": "value1"}`, Status: "Pending"},
		{ID: 2, Operation: "Update", TableName: "table2", UserID: 2, EntryID: "entry2", Data: `{"key2": "value2"}`, Status: "Pending"},
	}

	assert.Equal(t, expected, entries)
}

func TestUpdateSyncEntryStatus(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Вставляем тестовые данные в таблицу SyncQueue
	_, err := db.Exec("INSERT INTO SyncQueue (id, table_name, user_id, entry_id, operation, data, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "TestTable", 1, "TestEntryID", "Create", "Test data", "Pending")
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Вызываем метод UpdateSyncEntryStatus
	err = keeper.UpdateSyncEntryStatus(ctx, 1, "Done")
	if err != nil {
		t.Fatalf("UpdateSyncEntryStatus failed: %v", err)
	}

	// Проверяем, изменился ли статус записи в таблице SyncQueue
	var status string
	err = db.QueryRow("SELECT status FROM SyncQueue WHERE id = ?", 1).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query status: %v", err)
	}

	// Проверяем ожидаемый статус
	expectedStatus := "Done"
	if status != expectedStatus {
		t.Errorf("unexpected status: got %s, want %s", status, expectedStatus)
	}
}
func TestGetData(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Вставляем тестовые данные в таблицу
	_, err := db.Exec("INSERT INTO UserCredentials (id, user_id, login, password, meta_info, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"id_tab", 1, "value1", "value2", "value3", time.Now())
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Вызываем метод GetData
	data, err := keeper.GetData(ctx, "UserCredentials", 1, "id_tab")
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}

	// Ожидаемые данные
	expectedData := map[string]string{
		"login":     "value1",
		"password":  "value2",
		"meta_info": "value3",
	}

	// Проверяем, совпадают ли полученные данные с ожидаемыми
	for key, expectedValue := range expectedData {
		if value, ok := data[key]; !ok || value != expectedValue {
			t.Errorf("unexpected value for column %s: got %s, want %s", key, value, expectedValue)
		}
	}
}
func TestDeleteData(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Вставляем тестовые данные в таблицу
	_, err := db.Exec("INSERT INTO UserCredentials (id, user_id, login, password, meta_info, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"id_tab", 1, "value1", "value2", "value3", time.Now())
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Удаляем данные с указанным user_id и entry_id
	err = keeper.DeleteData(ctx, "UserCredentials", 1, "id_tab")
	if err != nil {
		t.Fatalf("DeleteData failed: %v", err)
	}

	// Проверяем, что запись была удалена
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM UserCredentials WHERE id = ?", 1).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query record count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected record count to be 0, got %d", count)
	}
}

func TestUpdateData(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Вставляем тестовые данные в таблицу
	_, err := db.Exec("INSERT INTO UserCredentials (id, user_id, login, password, meta_info, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"id_tab", 1, "value1", "value2", "value3", time.Now())
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Обновляем данные с указанным user_id и entry_id
	data := map[string]string{
		"login":     "updated_value1",
		"password":  "updated_value2",
		"meta_info": "updated_value3",
	}

	err = keeper.UpdateData(ctx, "UserCredentials", 1, "id_tab", data)
	if err != nil {
		t.Fatalf("UpdateData failed: %v", err)
	}

	// Проверяем, что запись была обновлена
	var login, password, metaInfo string
	err = db.QueryRow("SELECT login, password, meta_info FROM UserCredentials WHERE user_id = ? AND id = ?", 1, "id_tab").Scan(&login, &password, &metaInfo)
	if err != nil {
		t.Fatalf("failed to query record: %v", err)
	}

	if login != "updated_value1" || password != "updated_value2" || metaInfo != "updated_value3" {
		t.Errorf("expected updated values, got login: %s, password: %s, metaInfo: %s", login, password, metaInfo)
	}
}

func TestAddData(t *testing.T) {
	// Устанавливаем тестовую базу данных
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	ctx := context.Background()
	keeper := bdkeeper.NewKeeper(db)

	// Подготавливаем данные для вставки
	testData := map[string]string{
		"login":      "value1",
		"password":   "value2",
		"meta_info":  "value3",
		"updated_at": time.Now().Format(time.RFC3339),
	}

	// Задаем значения user_id и entry_id для теста
	user_id := 1
	entry_id := "entry_id_1"

	// Вызываем метод AddData
	err := keeper.AddData(ctx, "UserCredentials", user_id, entry_id, testData)
	if err != nil {
		t.Fatalf("AddData failed: %v", err)
	}

	// Проверяем, что данные были успешно добавлены, проверяя их присутствие в базе данных
	var login, password, meta_info string
	err = db.QueryRow("SELECT login, password, meta_info FROM UserCredentials WHERE user_id = ? AND id = ?", user_id, entry_id).Scan(&login, &password, &meta_info)
	if err != nil {
		t.Fatalf("failed to query record: %v", err)
	}

	// Проверяем, что данные в базе данных соответствуют ожидаемым данным
	expectedValue1 := "value1"
	expectedValue2 := "value2"
	expectedValue3 := "value3"
	if login != expectedValue1 || password != expectedValue2 || meta_info != expectedValue3 {
		t.Fatalf("Unexpected data retrieved from database. Got: %s, %s, %s; Want: %s, %s, %s",
			login, password, meta_info, expectedValue1, expectedValue2, expectedValue3)
	}
}

func TestNewKeeperWithMockDB(t *testing.T) {
	// Устанавливаем тестовую базу данных
	db, cleanup := setup(t)
	defer cleanup()

	// Создаем экземпляр Keeper
	keeper := bdkeeper.NewKeeper(db)

	// Проверяем, что экземпляр Keeper создан успешно
	if keeper == nil {
		t.Fatal("NewKeeper returned nil")
	}

}

func TestNewKeeperWithoutMockDB(t *testing.T) {

	// Создаем экземпляр Keeper
	keeper := bdkeeper.NewKeeper(nil)

	// Проверяем, что экземпляр Keeper создан успешно
	if keeper == nil {
		t.Fatal("NewKeeper returned nil")
	}

}
