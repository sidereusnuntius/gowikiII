package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/keystore"
	sqlkeystore "github.com/sidereusnuntius/gowiki/internal/keystore/sql"
)

func setupKeyStore(db *sql.DB) *keystore.KeyStore {
	sqlKeyStore := sqlkeystore.New(db)
	return &keystore.KeyStore{
		Store: sqlKeyStore,
	}
}
