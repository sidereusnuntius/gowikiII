package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
	"github.com/sidereusnuntius/gowiki/internal/validation"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
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
func (a *Auth) RegisterUser(ctx context.Context, in model.RegisterInput, admin bool) (model.Session, error) {
	normalizeFields(&in)
	err := validateUser(&in)
	if err != nil {
		return model.Session{}, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return model.Session{}, fmt.Errorf("failed to hash password: %w", err)
	}

	wikilog.Logger.Debug().Msgf("welcome aboard, %s!", in.Username)
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
		return model.Session{}, err
	}

	session := CreateSession(&user)
	if err = a.Sessions.SaveSession(ctx, session); err != nil {
		wikilog.Logger.Error().Err(err).Msg("a.Sessions.SaveSession(): failed to save session after registration")
		return model.Session{}, nil
	}

	return session, nil
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

// normalizeFields trims the username and email and turns them to lowercase.
func normalizeFields(in *model.RegisterInput) {
	in.Username = strings.ToLower(
		strings.TrimSpace(in.Username),
	)
	in.Email = strings.ToLower(
		strings.TrimSpace(in.Email),
	)
}

func validateUser(in *model.RegisterInput) error {
	ve := wikierr.NewValidationError()

	if err := validation.Username.Apply(in.Username); err != nil {
		ve.Append("username", err)
	}
	if err := validation.Email.Apply(in.Email); err != nil {
		ve.Append("email", err)
	}
	if err := validation.Password.Apply(in.Password); err != nil {
		ve.Append("password", err)
	}

	if len(ve.Fields) != 0 {
		return ve
	}

	return nil
}
