-- +goose Up
CREATE TABLE IF NOT EXISTS FilesData (
    id TEXT PRIMARY KEY,
    user_id INTEGER,
    path TEXT NOT NULL,
    meta_info TEXT,  
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES Users(id)
);

-- +goose Down
DROP TABLE FilesData;