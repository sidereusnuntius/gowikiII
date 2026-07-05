package tests

import (
	"context"
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

func TestDB(ctx context.Context) (*sql.DB, error) {
	config := config.DbConfig{
		URL: ":memory:",
	}

	return db.Open(ctx, config)
}
