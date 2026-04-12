package sqlkeystore

import (
	"context"
	"database/sql"

	sqlhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/sql"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type SqliteKeyStore struct {
	DB *sql.DB
}

func New(db *sql.DB) *SqliteKeyStore {
	return &SqliteKeyStore{
		DB: db,
	}
}

const (
	insertPublicKey  = "INSERT INTO public_keys (iri, owner_id, type, key_pem) VALUES (?, ?, ?, ?)"
	insertPrivateKey = "INSERT INTO private_keys (key_pem, owner_id, type) VALUES (?, ?, ?)"
	selectPublicKey  = `SELECT
		pk.id,
		pk.iri,
		owner_id,
		a.uri,
		pk.type,
		key_pem
		FROM public_keys pk
		LEFT JOIN actors a ON a.id = pk.owner_id
		WHERE pk.iri = ?
		LIMIT 1
		`
)

func (s *SqliteKeyStore) GetPublicKey(ctx context.Context, keyIRI string) (model.PublicKey, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectPublicKey, keyIRI)

	var (
		key        model.PublicKey
		pkType     sql.NullInt16
		pkOwnerId  sql.NullInt64
		pkOwnerIRI sql.NullString
	)
	err := row.Scan(
		&key.ID,
		&key.URI,
		&pkOwnerId,
		&pkOwnerIRI,
		&pkType,
		&key.Pem,
	)
	if err != nil {
		return model.PublicKey{}, sqlhelpers.HandleErr(err)
	}

	if pkOwnerId.Valid && pkOwnerIRI.Valid {
		key.OwnerID = pkOwnerId.Int64
		key.OwnerIRI = pkOwnerIRI.String
	}

	if pkType.Valid {
		key.Type = model.KeyType(pkType.Int16)
	}

	return key, nil
}

func (s *SqliteKeyStore) SavePublicKey(ctx context.Context, key *model.PublicKey) error {
	res, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx,
		insertPublicKey,
		key.URI,
		sqlhelpers.NullableInt64(key.OwnerID),
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
