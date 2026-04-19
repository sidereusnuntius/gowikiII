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
	insertPublicKey    = "INSERT INTO public_keys (iri, owner_id, type, key_pem) VALUES (?, ?, ?, ?)"
	insertPrivateKey   = "INSERT INTO private_keys (key_pem, owner_id, type) VALUES (?, ?, ?)"
	selectPublicKeyTpl = `SELECT
		pk.id,
		pk.iri,
		owner_id,
		a.uri,
		pk.type,
		key_pem
		FROM public_keys pk
		LEFT JOIN actors a ON a.id = pk.owner_id`
	selectPublicKeyByKeyIRI    = selectPublicKeyTpl + ` WHERE pk.iri = ? LIMIT 1`
	selectPrivateKeyByOwnerIRI = `SELECT
		k.id,
		k.key_pem,
		k.owner_id,
		k.type,
		pk.iri
		FROM private_keys k
		JOIN actors a ON k.owner_id = a.id
		JOIN public_keys pk ON pk.owner_id = a.id
		WHERE a.uri = ? LIMIT 1`
	publicKeyExists = "SELECT EXISTS(SELECT 1 FROM public_keys WHERE iri = ?)"
)

func (s *SqliteKeyStore) PublicKeyExists(ctx context.Context, keyIRI string) (bool, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, publicKeyExists, keyIRI)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, sqlhelpers.HandleErr(err)
	}

	return exists, nil
}

func (s *SqliteKeyStore) GetPublicKey(ctx context.Context, keyIRI string) (model.PublicKey, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectPublicKeyByKeyIRI, keyIRI)

	return scanPublicKey(row)
}

func scanPublicKey(row *sql.Row) (model.PublicKey, error) {
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

func (s *SqliteKeyStore) GetPrivateKey(ctx context.Context, ownerIRI string) (model.PrivateKey, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectPrivateKeyByOwnerIRI, ownerIRI)
	var (
		key     model.PrivateKey
		keyType sql.NullInt16
	)

	err := row.Scan(
		&key.ID,
		&key.Pem,
		&key.OwnerID,
		&keyType,
		&key.PublicKeyIRI,
	)
	if err != nil {
		return model.PrivateKey{}, sqlhelpers.HandleErr(err)
	}

	if keyType.Valid {
		key.Type = model.KeyType(keyType.Int16)
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
