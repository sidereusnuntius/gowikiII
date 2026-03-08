-- +goose Up
CREATE TABLE IF NOT EXISTS users (
       id INTEGER PRIMARY KEY,
       username VARCHAR(64) NOT NULL,
       email VARCHAR(255) NOT NULL,
       password BINARY NOT NULL,
       verified BOOLEAN DEFAULT FALSE NOT NULL,
       is_admin BOOLEAN DEFAULT FALSE NOT NULL,
       created_at INTEGER DEFAULT (unixepoch('now')) NOT NULL,

       UNIQUE (username),
       UNIQUE(email)
);

-- +goose Down
DROP TABLE users;
