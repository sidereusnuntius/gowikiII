package mocks

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockFedStore struct {
	mock.Mock
}

func (mfs *MockFedStore) GetHost(ctx context.Context, hostname string) (model.Host, error) {
	args := mfs.Called(ctx, hostname)
	return args.Get(0).(model.Host), args.Error(1)
}

func (mfs *MockFedStore) SaveHost(ctx context.Context, host *model.Host) error {
	args := mfs.Called(ctx, host)
	return args.Error(0)
}

func (mfs *MockFedStore) UpdateHostStatus(ctx context.Context, id int64, status model.HostStatus) error {
	args := mfs.Called(ctx)
	return args.Error(0)
}
