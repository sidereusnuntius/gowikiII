package keystore

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type KeyStore struct {
	Store Store
}

func New(store Store) KeyStore {
	return KeyStore{
		Store: store,
	}
}

// GenereteKeys creates a pair of public and private key for the given actor and stores them in the database. A URI is created for the public key based on the actor's URI.
func (ks *KeyStore) GenerateKeyPair(ctx context.Context, actorID int64, actorURI string, keyType model.KeyType) error {
	var (
		pubPem  []byte
		privPem []byte
		err     error
	)
	// Generate keys.
	// For now only RSA is supported.
	switch keyType {
	case model.RSAKey:
		pubPem, privPem, err = generateRSAKeyPair()
	default:
	}

	if err != nil {
		return err
	}

	// Generate public key URI.
	url, err := url.Parse(actorURI)
	if err != nil {
		return err
	}

	keyURI := url.JoinPath("main-key").String()

	pub := model.PublicKey{
		URI:     keyURI,
		OwnerID: actorID,
		Type:    keyType,
		Pem:     pubPem,
	}

	priv := model.PrivateKey{
		OwnerID: actorID,
		Type:    keyType,
		Pem:     privPem,
	}

	if err = ks.Store.SavePublicKey(ctx, &pub); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	if err = ks.Store.SavePrivateKey(ctx, &priv); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}
