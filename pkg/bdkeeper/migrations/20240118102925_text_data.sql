-- +goose Up
CREATE TABLE IF NOT EXISTS TextData (
    id TEXT PRIMARY KEY,
    user_id INTEGER,
    data TEXT NOT NULL,
    meta_info TEXT,
    deleted BOOLEAN DEFAULT FALSE,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES Users(id)
);

-- +goose Down
DROP TABLE TextData;