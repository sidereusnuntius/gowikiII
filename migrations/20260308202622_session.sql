-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
       token BINARY PRIMARY KEY,
       created_at INTEGER DEFAULT (unixepoch('now')) NOT NULL,
       expires_at INTEGER NOT NULL,
       user_id INTEGER NOT NULL,

       FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE INDEX IF NOT EXISTS idx_session_expiration ON sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_session_usr_expires ON sessions(user_id, expires_at);

-- +goose Down
DROP INDEX idx_session_usr_expires;
DROP INDEX idx_session_expiration;
DROP TABLE sessions;
