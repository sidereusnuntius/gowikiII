package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/valyala/fastjson"
)

type ApClient struct {
	Client *http.Client
}

func New() *ApClient {
	httpClient := &http.Client{}
	return &ApClient{
		Client: httpClient,
	}
}

func (ac *ApClient) FetchKey(ctx context.Context, keyId string) (model.PublicKey, error) {
	value, err := ac.Fetch(ctx, keyId)
	if err != nil {
		return model.PublicKey{}, err
	}

	key := value.Get("publicKey")
	if key == nil {
		return model.PublicKey{}, errors.New("key not found")
	}

	// TODO: what if owner is an object?
	keyIRI := key.GetStringBytes("id")
	owner := key.GetStringBytes("owner")
	keyPem := key.GetStringBytes("publicKeyPem")

	if len(keyIRI) == 0 || len(owner) == 0 || len(keyPem) == 0 {
		return model.PublicKey{}, errors.New("invalid public key")
	}

	pub := model.PublicKey{
		URI:      string(keyIRI),
		OwnerIRI: string(owner),
		Pem:      keyPem,
		Type:     model.RSAKey, // TODO: change this
	}

	return pub, nil
}

// TODO: sign GET requests.
func (ac *ApClient) Fetch(ctx context.Context, url string) (*fastjson.Value, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/ld+json")
	res, err := ac.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, errors.New("request failed") // TODO: handle response code and provide better message
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	value, err := fastjson.ParseBytes(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return value, nil
}
