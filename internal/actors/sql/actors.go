package actorstore

import (
	"context"
	"database/sql"
	"errors"
	"log"

	sqlhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/sql"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

const (
	insertActor = `INSERT INTO actors (
		user_id,
		uri,
		type,
		username,
		host,
		name,
		summary,

		inbox,
		outbox,
		followers,
		following,
		url,
		shared_inbox,

		published,
		updated
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING id`
	selectSharedInboxId = "SELECT id FROM shared_inboxes WHERE uri = ?"
	insertSharedInbox   = "INSERT INTO shared_inboxes (uri) VALUES (?) RETURNING id"
)

type ActorsStore struct {
	DB *sql.DB
}

func New(db *sql.DB) ActorsStore {
	return ActorsStore{
		DB: db,
	}
}

func (s *ActorsStore) CreateSharedInboxIfNotExists(ctx context.Context, uri string) (int64, error) {
	var id int64
	executor := txdb.GetExecutor(ctx, s.DB)

	row := executor.QueryRowContext(ctx, selectSharedInboxId, uri)
	if err := row.Scan(&id); err == nil {
		return id, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	row = executor.QueryRowContext(ctx, insertSharedInbox, uri)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *ActorsStore) SaveActor(ctx context.Context, actor *model.Actor) error {
	res, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx,
		insertActor,
		sql.NullInt64{
			Int64: actor.UserID,
			Valid: true,
		},
		actor.URI,
		actor.Type,
		actor.Username,
		actor.Host,
		sqlhelpers.NullableString(actor.DisplayName),
		sqlhelpers.NullableString(actor.Summary),
		sqlhelpers.NullableString(actor.Inbox),
		sqlhelpers.NullableString(actor.Outbox),
		sqlhelpers.NullableString(actor.Followers),
		sqlhelpers.NullableString(actor.Following),
		sqlhelpers.NullableString(actor.URL),
		sqlhelpers.NullableInt64(actor.SharedInboxID),
		sqlhelpers.NullableTimeUnix(actor.Published),
		sqlhelpers.NullableTimeUnix(actor.Updated),
	)
	if err != nil {
		return err // TODO: handle error.
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Print(err) // Return error?
	}
	actor.ID = id

	return nil
}
