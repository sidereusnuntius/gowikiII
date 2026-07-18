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
	selectActorByHandle   = selectActor + ` WHERE a.username = ? AND a.host = ? LIMIT 1`
	selectActorByIRI      = selectActor + " WHERE a.uri = ? LIMIT 1"
	selectActorByID       = selectActor + " WHERE a.id = ? LIMIT 1"
	selectActorIdByUserID = "SELECT id FROM actors WHERE user_id = ?"
	insertActor           = `INSERT INTO actors (
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
	insertFollow        = `INSERT INTO follows (
		iri,
		follower_id,
		followee_id,
		accepted,
		published
	) VALUES (?, ?, ?, ?, ?) RETURNING id`
	selectFollow = `SELECT
		f.id,
		f.iri,
		follower_id,
		follower.uri,
		followee_id,
		followee.uri,
		f.accepted,
		f.published
	FROM follows f
	JOIN actors follower ON follower.id = f.follower_id
	JOIN actors followee ON followee.id = f.followee_id
	WHERE iri = ?
	LIMIT 1
	`
	followSetAccepted = "UPDATE follows SET accepted = ? WHERE iri = ?"
	selectFollowers   = `SELECT
		followers.uri
	FROM follows
	JOIN actors followers ON follows.follower_id = followers.id
	JOIN actors followee ON follows.followee_id = followee.id
	WHERE followee.uri = ? AND follows.accepted`
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

func (s *ActorsStore) GetActorByID(ctx context.Context, id int64) (model.Actor, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectActorByID, id)

	return scanActor(row)
}

func (s *ActorsStore) GetActorIdForUser(ctx context.Context, userID int64) (int64, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectActorIdByUserID, userID)
	var id int64

	if err := row.Scan(&id); err != nil {
		return 0, sqlhelpers.HandleErr(err)
	}

	return id, nil
}

func (s *ActorsStore) GetActorByIRI(ctx context.Context, iri string) (model.Actor, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectActorByIRI, iri)

	return scanActor(row)
}

func (s *ActorsStore) SaveFollow(ctx context.Context, follow *model.Follow) error {
	result, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx, insertFollow,
		follow.IRI,
		follow.FollowerID,
		follow.FolloweeID,
		follow.Accepted,
		follow.Published.Unix(),
	)
	if err != nil {
		return sqlhelpers.HandleErr(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return sqlhelpers.HandleErr(err)
	}

	follow.ID = id
	return nil
}

func (s *ActorsStore) GetFollow(ctx context.Context, iri string) (model.Follow, error) {
	row := txdb.GetExecutor(ctx, s.DB).QueryRowContext(ctx, selectFollow, iri)

	var (
		follow       model.Follow
		acceptedBool sql.NullBool
		published    int64
	)

	err := row.Scan(
		&follow.ID,
		&follow.IRI,
		&follow.FollowerID,
		&follow.FollowerIRI,
		&follow.FolloweeID,
		&follow.FolloweeIRI,
		&acceptedBool,
		&published,
	)
	if err != nil {
		return model.Follow{}, sqlhelpers.HandleErr(err)
	}

	if acceptedBool.Valid {
		follow.Accepted = acceptedBool.Bool
	}

	follow.Published = time.Unix(published, 0)
	return follow, nil
}

func (s *ActorsStore) SetFollowAccepted(ctx context.Context, iri string, accepted bool) error {
	_, err := txdb.GetExecutor(ctx, s.DB).ExecContext(ctx, followSetAccepted, accepted, iri)
	return sqlhelpers.HandleErr(err)
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

func (s *ActorsStore) GetFollowers(ctx context.Context, actorIRI string) ([]string, error) {
	rows, err := txdb.GetExecutor(ctx, s.DB).QueryContext(ctx, selectFollowers, actorIRI)
	if err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	var (
		iris = make([]string, 0)
		curr string
	)
	for rows.Next() {
		if err = rows.Scan(&curr); err != nil {
			return nil, sqlhelpers.HandleErr(err)
		}

		iris = append(iris, curr)
	}

	if err = rows.Err(); err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	return iris, nil
}
