-- +goose Up
CREATE TABLE IF NOT EXISTS CreditCardData (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    card_number TEXT NOT NULL,
    expiration_date TEXT NOT NULL,
    cvv INTEGER NOT NULL,
    meta_info TEXT,
    deleted BOOLEAN DEFAULT FALSE,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES Users(id)
);

-- +goose Down
DROP TABLE CreditCardData;
