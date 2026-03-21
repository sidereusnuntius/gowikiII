package auth

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Store interface {
	// CreateUser stores the provided user, setting the Id field of the given
	// struct to the user's auto generated id.
	CreateUser(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	GetUserMinimalByEmail(ctx context.Context, email string) (model.User, error)
}

type SessionStore interface {
	GetFullSession(ctx context.Context, token string) (model.Session, error)
	SaveSession(ctx context.Context, session model.Session) error
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type ActorService interface {
	CreateLocalActor(ctx context.Context, username string, userID int64) error
}
