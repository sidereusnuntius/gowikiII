package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
	"github.com/sidereusnuntius/gowiki/migrations"
)

func Open(ctx context.Context, config config.DbConfig) (handle *sql.DB, err error) {
	url := config.URL + "?_journal=WAL&_timeout=5000&_fk=true" // TODO: improve this
	handle, err = sql.Open("sqlite3", url)
	if err != nil {
		err = fmt.Errorf("failed to open database connection: %w", err)
		return
	}

	defer func() {
		if err != nil {
			if err := handle.Close(); err != nil {
				wikilog.Logger.Error().Err(err).Msg("failed to close database connection after failure during setup")
			}
		}
	}()

	if err = handle.Ping(); err != nil {
		err = fmt.Errorf("failed to ping database: %w", err)
		return
	}

	// Run migrations, if needed.
	goose.SetLogger(wikilog.Logger)
	goose.SetBaseFS(migrations.MigrationsFS)
	if err = goose.SetDialect("sqlite3"); err != nil {
		err = fmt.Errorf("goose.SetDialect(): %w", err)
	}

	// TODO: set zerolog as goose's logger
	err = goose.UpContext(ctx, handle, ".")
	if err != nil {
		err = fmt.Errorf("goose.UpContext(): failed to apply migrations")
	}

	return
}
