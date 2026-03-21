package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionDuration    time.Duration = 24 * time.Hour
	SessionTokenLength               = 64
)

type Auth struct {
	TxManager *txdb.TxManager
	Actors    ActorService
	Store     Store
	Sessions  SessionStore
	// TODO: add email verification service etc.
}

func NewAuth(store Store, sessionStore SessionStore, manager *txdb.TxManager) *Auth {
	return &Auth{
		Store:     store,
		Sessions:  sessionStore,
		TxManager: manager,
	}
}

// TODO: add logic for creating ActivityPub actor and email verification.
func (a *Auth) RegisterUser(ctx context.Context, in model.RegisterInput, admin bool) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	log.Debug().Str("email", in.Email).Str("username", in.Username).Msg("registering user")
	// Perform validation.
	user := model.User{
		Username: in.Username,
		Email:    in.Email,
		Password: hashed,
		IsAdmin:  admin,
		Verified: admin, // Admin account need not be verified.
	}

	err = a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		err := a.Store.CreateUser(ctx, &user)
		if err != nil {
			return err
		}

		if err = a.Actors.CreateLocalActor(ctx, user.Username, user.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) Authenticate(ctx context.Context, in model.LoginInput) (model.Session, error) {
	user, err := a.Store.GetUserMinimalByEmail(ctx, in.Email)
	if err != nil {
		return model.Session{}, fmt.Errorf("failed to find user on database: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(in.Password))
	if err != nil {
		return model.Session{}, fmt.Errorf("wrong password")
	}

	session := CreateSession(&user)
	if err = a.Sessions.SaveSession(ctx, session); err != nil {
		return model.Session{}, err
	}

	return session, nil
}

func CreateSession(user *model.User) model.Session {
	buffer := make([]byte, SessionTokenLength)
	_, _ = rand.Read(buffer)

	return model.Session{
		Token:      hex.EncodeToString(buffer),
		Expiration: time.Now().Add(SessionDuration),
		User:       *user,
	}
}
