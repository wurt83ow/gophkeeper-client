package models

type SyncQueue struct {
	ID        int    `db:"id"`
	TableName string `db:"table_name"`
	UserID    int    `db:"user_id"`
	EntryID   string `db:"entry_id"`
	Operation string `db:"operation"`
	Data      string `db:"data"`
	Status    string `db:"status"`
}
