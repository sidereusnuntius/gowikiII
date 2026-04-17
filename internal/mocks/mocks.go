// This package contains mocks shared by multiple packages.
package mocks

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
)

type MockPublisher struct {
	mock.Mock
}

func (mp *MockPublisher) Publish(ctx context.Context, job river.JobArgs) error {
	args := mp.Called(ctx, job)

	return args.Error(0)
}
