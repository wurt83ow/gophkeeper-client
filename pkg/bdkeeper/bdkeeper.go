package bdkeeper

import (
	"database/sql"

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

func (k *Keeper) AddData(key string, value string) {
	stmt, _ := k.db.Prepare("INSERT INTO data(key, value) values(?,?)")
	stmt.Exec(key, value)
}

func (k *Keeper) UpdateData(key string, value string) {
	stmt, _ := k.db.Prepare("UPDATE data SET value = ? WHERE key = ?")
	stmt.Exec(value, key)
}

func (k *Keeper) DeleteData(key string) {
	stmt, _ := k.db.Prepare("DELETE FROM data WHERE key = ?")
	stmt.Exec(key)
}

func (k *Keeper) GetData(key string) string {
	row := k.db.QueryRow("SELECT value FROM data WHERE key = ?", key)
	var value string
	row.Scan(&value)
	return value
}

func (k *Keeper) GetAllData() map[string]string {
	rows, _ := k.db.Query("SELECT key, value FROM data")
	defer rows.Close()

	data := make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		rows.Scan(&key, &value)
		data[key] = value
	}
	return data
}

func (k *Keeper) ClearData() {
	stmt, _ := k.db.Prepare("DELETE FROM data")
	stmt.Exec()
}
