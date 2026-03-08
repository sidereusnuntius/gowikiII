package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteStore struct {
	DB *sql.DB
}

func Init(db *sql.DB) (*SqliteStore, error) {
	store := SqliteStore{
		DB: db,
	}

	return &store, nil
}
