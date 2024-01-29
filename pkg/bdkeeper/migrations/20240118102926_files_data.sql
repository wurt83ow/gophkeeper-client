-- +goose Up
CREATE TABLE IF NOT EXISTS FilesData (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    path TEXT NOT NULL,
    meta_info TEXT,
    deleted BOOLEAN DEFAULT FALSE,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES Users(id)
);

-- +goose Down
DROP TABLE FilesData;
