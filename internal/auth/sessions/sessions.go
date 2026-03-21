package sessions

import (
	"context"
	"database/sql"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type SqliteSessionStore struct {
	DB *sql.DB
}

func New(db *sql.DB) SqliteSessionStore {
	return SqliteSessionStore{
		DB: db,
	}
}

const (
	selectFullSession     = "SELECT s.token, s.expires_at, s.created_at, u.username, u.email, u.verified, u.is_admin FROM sessions s JOIN users u ON s.user_id = u.id WHERE s.token = ?"
	saveSession           = "INSERT INTO sessions (token, expires_at, user_id) VALUES (?, ?, ?)"
	deleteSession         = "DELETE FROM sessions WHERE token = ?"
	deleteExpiredSessions = "DELETE FROM sessions WHERE expires_at <= (unixepoch('now'))"
)

func (s *SqliteSessionStore) GetFullSession(ctx context.Context, token string) (model.Session, error) {
	var (
		created int64
		expires int64
	)
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectFullSession, token)
	var session model.Session
	err := row.Scan(
		&session.Token,
		&expires,
		&created,
		&session.User.Username,
		&session.User.Email,
		&session.User.Verified,
		&session.User.IsAdmin,
	)
	if err != nil {
		return model.Session{}, err
	}

	session.Expiration = time.Unix(expires, 0)
	session.Created = time.Unix(created, 0)

	return session, nil
}

func (s *SqliteSessionStore) SaveSession(ctx context.Context, session model.Session) error {
	_, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx, saveSession, session.Token, session.Expiration.Unix(), session.User.ID)
	return err
}

func (s *SqliteSessionStore) DeleteSession(ctx context.Context, token string) error {
	_, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx, deleteSession, token)
	return err
}

func (s *SqliteSessionStore) DeleteExpiredSessions(ctx context.Context) error {
	_, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx, deleteExpiredSessions)
	return err
}
