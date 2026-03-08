package db

import (
	"context"
	"github.com/sidereusnuntius/gowiki/internal/model"
)

type UserStore interface {
	// CreateUser stores the provided user, setting the Id field of the given
	// struct to the user's auto generated id.
	CreateUser(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
