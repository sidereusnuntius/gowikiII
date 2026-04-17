package mocks

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/stretchr/testify/mock"
)

type MockActors struct {
	mock.Mock
}

func (ma *MockActors) ActorExists(ctx context.Context, id string) (bool, error) {
	args := ma.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (ma *MockActors) CacheRemoteActor(ctx context.Context, actor activitystreams.Actor) error {
	args := ma.Called(ctx, actor)
	return args.Error(0)
}

func (ma *MockActors) CreateLocalActor(ctx context.Context, username string, userID int64) error {
	args := ma.Called(ctx, username, userID)
	return args.Error(0)
}

func (ma *MockActors) GetLocalActor(ctx context.Context, username string) (model.Actor, error) {
	args := ma.Called(ctx, username)
	return args.Get(0).(model.Actor), args.Error(1)
}
