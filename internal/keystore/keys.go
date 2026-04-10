package keystore

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/httpsig"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

type Fetcher interface {
	FetchKey(ctx context.Context, keyId string) (model.PublicKey, error)
}

type KeyStore struct {
	Store      Store
	KeyFetcher Fetcher
}

func New(store Store, fetcher Fetcher) KeyStore {
	return KeyStore{
		Store:      store,
		KeyFetcher: fetcher,
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

func (ks *KeyStore) VerifySignature(ctx context.Context, r *http.Request) error {
	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		return fmt.Errorf("httpsig.NewVerifier(): %w", err)
	}

	keyId := verifier.KeyId()
	keyIdUrl, _ := url.Parse(keyId)
	if keyIdUrl.Host == config.Config.Host {
		return errors.New("invalid activity origin")
	}

	key, err := ks.Store.GetPublicKey(ctx, keyId)
	if err != nil {
		if !errors.Is(err, wikierr.ErrNotFound) {
			return err
		}

		// If key does not exist in local db, we try to fetch it.
		key, err = ks.KeyFetcher.FetchKey(ctx, keyId)
		if err != nil {
			return err
		}

		if err = ks.Store.SavePublicKey(ctx, &key); err != nil {
			return err
		}
	}

	block, _ := pem.Decode(key.Pem)

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	var algorithm httpsig.Algorithm

	switch pub.(type) {
	case *rsa.PublicKey:
		algorithm = httpsig.RSA_SHA256 // TODO: we also have to get the digest algorithm.
	case *ecdsa.PublicKey:
		algorithm = httpsig.ECDSA_SHA256
	}

	return verifier.Verify(pub, algorithm)
}
