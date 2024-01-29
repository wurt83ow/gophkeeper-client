-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE Users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    deleted BOOLEAN DEFAULT FALSE
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE Users;
