package sqlkeystore

import (
	"context"
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type SqliteKeyStore struct {
	DB *sql.DB
}

func New(db *sql.DB) SqliteKeyStore {
	return SqliteKeyStore{
		DB: db,
	}
}

const (
	insertPublicKey  = "INSERT INTO public_keys (iri, owner_id, type, key_pem) VALUES (?, ?, ?, ?)"
	insertPrivateKey = "INSERT INTO private_keys (key_pem, owner_id, type) VALUES (?, ?, ?)"
)

func (s *SqliteKeyStore) SavePublicKey(ctx context.Context, key *model.PublicKey) error {
	res, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx,
		insertPublicKey,
		key.URI,
		key.OwnerID,
		int(key.Type),
		key.Pem,
	)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err // TODO: just log the error
	}

	key.ID = id
	return nil
}

func (s *SqliteKeyStore) SavePrivateKey(ctx context.Context, key *model.PrivateKey) error {
	_, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx,
		insertPrivateKey,
		key.Pem,
		key.OwnerID,
		int(key.Type),
	)
	return err // TODO: handle error
}
