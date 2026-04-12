package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
)

type Client interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
	FetchKey(ctx context.Context, keyId string) (model.PublicKey, error)
}

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
	res, err := ac.Fetch(ctx, keyId)
	if err != nil {
		return model.PublicKey{}, err
	}

	obj, err := activitystreams.ReadObject(res)
	if err != nil {
		return model.PublicKey{}, fmt.Errorf("activitystreams.ReadObject(): %w", err)
	}

	return obj.PublicKey()
}

// TODO: sign GET requests.
func (ac *ApClient) Fetch(ctx context.Context, url string) ([]byte, error) {
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

	return body, nil
}
