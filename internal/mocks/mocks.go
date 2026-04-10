// This package contains mocks shared by multiple packages.
package mocks

import (
	"context"

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
