package keystore

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"
	"time"

	"code.superseriousbusiness.org/httpsig"
	"github.com/sidereusnuntius/gowiki/internal/mocks"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/stretchr/testify/mock"
)

// Mock store on which we store keys locally. Perhaps rename KeyStore into SignatureService, Security or something like that?
type mockStore struct {
	mock.Mock
}

func (ms *mockStore) GetPublicKey(ctx context.Context, keyIRI string) (model.PublicKey, error) {
	args := ms.MethodCalled("GetPublicKey", ctx, keyIRI)
	key, _ := args.Get(0).(model.PublicKey)
	return key, args.Error(1)
}

func (ms *mockStore) SavePublicKey(ctx context.Context, key *model.PublicKey) error {
	args := ms.Called(ctx, key)
	return args.Error(0)
}

func (ms *mockStore) SavePrivateKey(ctx context.Context, key *model.PrivateKey) error {
	args := ms.Called(ctx, key)
	return args.Error(0)
}

var (
	store     = new(mockStore)
	client    = new(mocks.MockClient)
	prefs     = []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlg = httpsig.DigestSha256
	sigScheme = httpsig.Signature
	expiresIn = 10 * time.Second
	headers   = []string{httpsig.RequestTarget, "date", "digest"}
)

func initialize() (KeyStore, *mockStore, *mocks.MockClient) {
	store := new(mockStore)
	client := new(mocks.MockClient)

	ks := New(store, client)
	return ks, store, client
}

func request(t *testing.T, url string, body []byte) *http.Request {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("signedRequest(): failed to create request: %v", err)
	}

	req.Header.Add("Date", time.Now().Format(http.TimeFormat))

	return req
}

func signedRequest(t *testing.T, req *http.Request, body, privateKey []byte, pubKeyId string) *http.Request {
	block, _ := pem.Decode(privateKey)
	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("signedRequest(): failed to parse private key: %v", err)
	}

	key := keyAny.(*rsa.PrivateKey)

	signer, _, err := httpsig.NewSigner(prefs, digestAlg, headers, sigScheme, int64(expiresIn))
	if err != nil {
		t.Fatalf("signedRequest(): failed to create signer: %v", err)
	}

	if err = signer.SignRequest(key, pubKeyId, req, body); err != nil {
		t.Fatalf("signedRequest(): failed to sign request: %v", err)
	}

	return req
}

func createPublicKey(username string, pem []byte) model.PublicKey {
	actorId := "https://test.wiki/u/" + username
	keyId := actorId + "/main-key"
	return model.PublicKey{
		ID:       1,
		URI:      keyId,
		OwnerID:  1,
		OwnerIRI: actorId,
		Type:     model.RSAKey,
		Pem:      pem,
	}
}

func TestVerifySignature(t *testing.T) {
	// Generate keys for Alice
	pubAlicePem, privAlice, _ := generateRSAKeyPair()
	pubAlice := createPublicKey("alice", pubAlicePem)
	helloWorld := []byte("hello, world!")

	pubTrudyPem, privTrudy, _ := generateRSAKeyPair()
	_ = createPublicKey("alice", pubTrudyPem)

	// body := []byte("hello, world!")
	// req := signedRequest(t, "https://bio.wiki/inbox", body, priv, publicKey.URI)

	cases := []struct {
		title      string
		publicKey  model.PublicKey
		privateKey []byte
		signed     bool
		req        *http.Request
		shouldFail bool
		// error returned by call to local store; can be nil
		storeErr error
		// whether the key should be fetched via HTTP
		shouldFetch bool
		// error when fetching public key via client
		fetchErr error
		saveErr  error
	}{
		{
			title:      "signed request with key cached in store",
			publicKey:  pubAlice,
			privateKey: privAlice,
			signed:     true,
			req: signedRequest(
				t,
				request(t, "https://bio.wiki/inbox", helloWorld),
				helloWorld,
				privAlice,
				pubAlice.URI,
			),
			shouldFail:  false,
			storeErr:    nil,
			shouldFetch: false,
			fetchErr:    nil,
			saveErr:     nil,
		},
		{
			title:      "signed request with key that needs to be fetched",
			publicKey:  pubAlice,
			privateKey: privAlice,
			signed:     true,
			req: signedRequest(
				t,
				request(t, "https://bio.wiki/inbox", helloWorld),
				helloWorld,
				privAlice,
				pubAlice.URI,
			),
			shouldFail:  false,
			storeErr:    wikierr.ErrNotFound,
			shouldFetch: true,
			fetchErr:    nil,
			saveErr:     nil,
		},
		{
			title:       "request not signed",
			publicKey:   pubAlice,
			privateKey:  privAlice,
			signed:      false,
			req:         request(t, "https://bio.wiki/inbox", helloWorld),
			shouldFail:  true,
			storeErr:    nil,
			shouldFetch: true,
			fetchErr:    nil,
			saveErr:     nil,
		},
		{
			title:      "signed request with wrong key (cached)",
			publicKey:  pubAlice,
			privateKey: privTrudy,
			signed:     true,
			req: signedRequest(
				t,
				request(t, "https://bio.wiki/inbox", helloWorld),
				helloWorld,
				privTrudy,
				pubAlice.URI,
			),
			shouldFail:  true,
			storeErr:    nil,
			shouldFetch: false,
			fetchErr:    nil,
			saveErr:     nil,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			ks, store, client := initialize()

			var storeReturn model.PublicKey
			if c.storeErr == nil {
				storeReturn = c.publicKey
			}

			store.On("GetPublicKey", mock.Anything, c.publicKey.URI).Return(storeReturn, c.storeErr)

			// If the key is not cached locally, fetch it and save it.
			if c.shouldFetch {
				var fetchReturn any
				if c.fetchErr == nil {
					fetchReturn = c.publicKey
				}
				client.On("FetchKey", mock.Anything, c.publicKey.URI).Return(fetchReturn, c.fetchErr)
				store.On("SavePublicKey", mock.Anything, &c.publicKey).Return(c.saveErr)
			}

			err := ks.VerifySignature(t.Context(), c.req)
			if err != nil && !c.shouldFail {
				t.Fatalf("ks.VerifySignature(): unexpected error: %v", err)
			}

			if !c.signed {
				return
			}

			store.AssertCalled(t, "GetPublicKey", mock.Anything, c.publicKey.URI)
			if c.shouldFetch {
				client.AssertCalled(t, "FetchKey", mock.Anything, c.publicKey.URI)

				if c.fetchErr == nil {
					store.AssertCalled(t, "SavePublicKey", mock.Anything, &c.publicKey)
				} else {
					store.AssertNotCalled(t, "SavePublicKey", mock.Anything, &c.publicKey)
				}
			} else {
				client.AssertNotCalled(t, "FetchKey")
			}
		})
	}
}

// // Public key is cache in local database; use it, and don't fetch it from remote server.
// func TestVerifySignature_CachedInDb(t *testing.T) {
// 	ks, store, client := initialize()

// 	err := ks.VerifySignature(t.Context(), req)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	client.AssertNotCalled(t, "FetchKey")
// 	store.AssertCalled(t, "GetPublicKey", mock.Anything, publicKey.URI)
// }
