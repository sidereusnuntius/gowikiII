package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/security"
	sqlkeystore "github.com/sidereusnuntius/gowiki/internal/security/sql"
)

func setupSecurity(config config.WikiConfig, db *sql.DB, client security.Client) (*security.Security, error) {
	sqlKeyStore := sqlkeystore.New(db)
	return security.New(config, sqlKeyStore, client)
}
