package mocks

import (
	"context"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (mc *MockClient) FetchKey(ctx context.Context, keyId string) (model.PublicKey, error) {
	args := mc.Called(ctx, keyId)
	return args.Get(0).(model.PublicKey), args.Error(1)
}

func (mc *MockClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	args := mc.Called(ctx, url)
	return args.Get(0).([]byte), args.Error(1)
}

func (mc *MockClient) Post(ctx context.Context, r *http.Request) error {
	args := mc.Called(ctx, r)
	return args.Error(0)
}
