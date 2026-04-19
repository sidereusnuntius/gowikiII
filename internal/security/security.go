package security

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.superseriousbusiness.org/httpsig"
	"github.com/goccy/go-json"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

var algorithmPreferences = []httpsig.Algorithm{httpsig.RSA_SHA256}
var headers = []string{httpsig.RequestTarget, "date", "digest"}

const (
	digestAlgorithm = httpsig.DigestSha256
	expiresIn       = 1 * time.Minute
)

type Client interface {
	FetchKey(ctx context.Context, keyId string) (model.PublicKey, error)
	Post(ctx context.Context, r *http.Request) error
}

type Security struct {
	Store  Store
	Client Client
	Config config.WikiConfig
	// TODO: change this to a pool.
	signer      httpsig.Signer
	signerMutex sync.Mutex
}

func New(config config.WikiConfig, store Store, client Client) (*Security, error) {
	signer, _, err := httpsig.NewSigner(algorithmPreferences, digestAlgorithm, headers, httpsig.Authorization, int64(expiresIn))
	if err != nil {
		return nil, err
	}
	return &Security{
		Store:  store,
		Client: client,
		Config: config,
		signer: signer,
	}, nil
}

func (s *Security) SavePublicKey(ctx context.Context, key *model.PublicKey) error {
	// TODO: validation.
	return s.Store.SavePublicKey(ctx, key)
}

func (s *Security) PublicKeyExists(ctx context.Context, keyIRI string) (bool, error) {
	return s.Store.PublicKeyExists(ctx, keyIRI)
}

// GenereteKeys creates a pair of public and private key for the given actor and stores them in the database. A URI is created for the public key based on the actor's URI.
func (s *Security) GenerateKeyPair(ctx context.Context, actorID int64, actorURI string, keyType model.KeyType) error {
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

	if err = s.Store.SavePublicKey(ctx, &pub); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	if err = s.Store.SavePrivateKey(ctx, &priv); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}

func (s *Security) VerifySignature(ctx context.Context, r *http.Request) error {
	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		return fmt.Errorf("httpsig.NewVerifier(): %w", err)
	}

	keyId := verifier.KeyId()
	keyIdUrl, _ := url.Parse(keyId)
	if keyIdUrl.Host == s.Config.Host {
		return errors.New("invalid activity origin")
	}

	key, err := s.Store.GetPublicKey(ctx, keyId)
	if err != nil {
		if !wikierr.Is(err, wikierr.ErrNotFound) {
			return err
		}

		// If key does not exist in local db, we try to fetch it.
		key, err = s.Client.FetchKey(ctx, keyId)
		if err != nil {
			return err
		}

		if err = s.Store.SavePublicKey(ctx, &key); err != nil {
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

func (s *Security) signedRequest(ctx context.Context, inbox string, body []byte, actorIRI string) (*http.Request, error) {
	s.signerMutex.Lock()
	defer s.signerMutex.Unlock()

	privKey, err := s.Store.GetPrivateKey(ctx, actorIRI)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, inbox, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Add("Content-Type", "application/ld+json")

	block, _ := pem.Decode(privKey.Pem)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	if err = s.signer.SignRequest(key, privKey.PublicKeyIRI, r, body); err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Security) PostSigned(ctx context.Context, inbox string, payload any, actorIRI string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	r, err := s.signedRequest(ctx, inbox, body, actorIRI)
	if err != nil {
		return err
	}
	return s.Client.Post(ctx, r)

	// TODO: consider the type of the key and user an appropriate signer.

}
