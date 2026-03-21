package authstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

const (
	insertUser            = "INSERT INTO users (username, email, password, verified, is_admin) VALUES (?, ?, ?, ?, ?) RETURNING id"
	existsByEmail         = "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)"
	existsByUsername      = "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)"
	getUserMinimalByEmail = "SELECT id, username, password FROM users where email = ?"
)

type AuthStore struct {
	DB *sql.DB
}

func New(db *sql.DB) AuthStore {
	return AuthStore{
		DB: db,
	}
}

func (s *AuthStore) CreateUser(ctx context.Context, user *model.User) error {
	res, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx,
		insertUser,
		user.Username,
		user.Email,
		user.Password,
		user.Verified,
		user.IsAdmin,
	)
	if err != nil {
		// TODO: handle error
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user's generated id: %w", err)
	}

	user.ID = id
	return nil
}

func (s *AuthStore) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, existsByEmail, email)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	// TODO: treat error.
	return exists, nil
}

func (s *AuthStore) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, existsByUsername, username)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		// TODO: treat error
		return false, err
	}
	return exists, nil
}

func (s *AuthStore) GetUserMinimalByEmail(ctx context.Context, email string) (model.User, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, getUserMinimalByEmail, email)
	var user model.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		return model.User{}, err
	}

	return user, nil
}
