package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/auth/sessions"
)

func setupSessionsStore(db *sql.DB) *sessions.SqliteSessionStore {
	return &sessions.SqliteSessionStore{
		DB: db,
	}
}
