-- +goose Up
CREATE TABLE IF NOT EXISTS CreditCardData (
    id TEXT PRIMARY KEY,
    user_id INTEGER,
    card_number TEXT NOT NULL,
    expiration_date TEXT NOT NULL,
    cvv TEXT NOT NULL,
    meta_info TEXT,   
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES Users(id)
);

-- +goose Down
DROP TABLE CreditCardData;