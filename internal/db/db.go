package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/config"
)

func Open(config config.DbConfig) (*sql.DB, error) {
	url := config.URL + "?_journal=WAL&_timeout=5000&_fk=true"
	pool, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(); err != nil {
		if err2 := pool.Close(); err2 != nil {
			log.Error().Err(err2).Msg("failed to close connection connection after failing to ping database")
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
