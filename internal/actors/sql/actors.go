package actorstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sqlhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/sql"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

const (
	selectActor = `SELECT
		a.id,
		a.uri,
		a.type,
		a.username,
		a.host,
		a.name,
		a.summary,

		a.inbox,
		a.outbox,
		a.followers,
		a.following,
		a.url,
		a.published,
		a.updated,

		a.shared_inbox,
		si.uri,

		pk.iri,
		pk.type,
		pk.key_pem
	FROM actors a
	LEFT JOIN public_keys pk ON pk.owner_id = a.id
	LEFT JOIN shared_inboxes si ON si.id = a.shared_inbox
	`
	selectActorByHandle = selectActor + ` WHERE a.username = ? AND a.host = ? LIMIT 1`
	selectActorByIRI    = selectActor + " WHERE a.uri = ? LIMIT 1"
	insertActor         = `INSERT INTO actors (
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
	actorExists         = "SELECT EXISTS(SELECT 1 FROM actors WHERE uri = ?)"
)

type ActorsStore struct {
	DB *sql.DB
}

func New(db *sql.DB) ActorsStore {
	return ActorsStore{
		DB: db,
	}
}

func (s *ActorsStore) ActorExists(ctx context.Context, iri string) (bool, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, actorExists, iri)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, sqlhelpers.HandleErr(err)
	}

	return exists, nil
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
		sqlhelpers.NullableInt64(actor.UserID),
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
		return err // Return error?
	}
	actor.ID = id

	return nil
}

func (s *ActorsStore) GetActorByHandle(ctx context.Context, username, host string) (model.Actor, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectActorByHandle,
		username,
		host,
	)

	return scanActor(row)
}

func (s *ActorsStore) GetActorByIRI(ctx context.Context, iri string) (model.Actor, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectActorByIRI, iri)

	return scanActor(row)
}

func scanActor(row *sql.Row) (model.Actor, error) {
	var (
		actor                                    model.Actor
		displayName, summary                     sql.NullString
		inbox, outbox, followers, following, url sql.NullString
		published, updated                       sql.NullInt64
		sharedInboxID                            sql.NullInt64
		sharedInboxIRI, pkIri                    sql.NullString
		pkType                                   sql.NullInt16
		pkPem                                    []byte
	)
	err := row.Scan(
		&actor.ID,
		&actor.URI,
		&actor.Type,
		&actor.Username,
		&actor.Host,

		&displayName,
		&summary,

		&inbox,
		&outbox,
		&followers,
		&following,
		&url,

		&published,
		&updated,

		&sharedInboxID,
		&sharedInboxIRI,

		&pkIri,
		&pkType,
		&pkPem,
	)

	if err != nil {
		return model.Actor{}, sqlhelpers.HandleErr(err)
	}

	if displayName.Valid {
		actor.DisplayName = displayName.String
	}
	if summary.Valid {
		actor.Summary = summary.String
	}

	if inbox.Valid {
		actor.Inbox = inbox.String
	}
	if outbox.Valid {
		actor.Outbox = outbox.String
	}
	if followers.Valid {
		actor.Followers = followers.String
	}
	if following.Valid {
		actor.Following = following.String
	}
	if sharedInboxID.Valid && sharedInboxIRI.Valid {
		actor.SharedInboxID = sharedInboxID.Int64
		actor.SharedInbox = sharedInboxIRI.String
	}

	if published.Valid {
		actor.Published = time.Unix(published.Int64, 0)
	}
	if updated.Valid {
		actor.Updated = time.Unix(updated.Int64, 0)
	}

	if pkIri.Valid && pkType.Valid && len(pkPem) > 0 {
		pk := model.PublicKey{
			URI:  pkIri.String,
			Type: model.KeyType(pkType.Int16),
			Pem:  pkPem,
		}

		actor.PublicKey = pk
	}

	return actor, nil
}
