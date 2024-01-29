-- +goose Up
CREATE TABLE IF NOT EXISTS SyncQueue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    table_name TEXT NOT NULL,
    data TEXT NOT NULL,
    file_path TEXT,
    deleted BOOLEAN DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS SyncQueue;