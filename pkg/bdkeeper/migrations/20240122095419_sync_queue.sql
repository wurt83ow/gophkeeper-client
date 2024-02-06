-- +goose Up
CREATE TABLE IF NOT EXISTS SyncQueue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    table_name TEXT NOT NULL,
    user_id INTEGER,  
    entry_id TEXT,
    operation TEXT NOT NULL CHECK(operation IN ('Create', 'Update', 'Delete')),
    data TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('Pending', 'Progress', 'Done', 'Error'))   
);

-- +goose Down
DROP TABLE IF EXISTS SyncQueue;