package wiki

import (
	"context"
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/transactions"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	TxManager *txdb.TxManager
	Store     db.UserStore
	// TODO: add email verification service etc.
}

func NewAuth(store db.UserStore, manager *txdb.TxManager) *Auth {
	return &Auth{
		Store:     store,
		TxManager: manager,
	}
}

// TODO: add logic for creating ActivityPub actor and email verification.
func (a *Auth) RegisterUser(ctx context.Context, in model.RegisterInput, admin bool) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Perform validation.
	user := model.User{
		Username: in.Username,
		Email:    in.Email,
		Password: hashed,
		IsAdmin:  admin,
		Verified: admin, // Admin account does not need to be verified.
	}

	err = a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		err := a.Store.CreateUser(ctx, &user)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
