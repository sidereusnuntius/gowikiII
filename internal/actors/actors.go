package actors

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/sidereusnuntius/gowiki/internal/model/streams"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	"github.com/sidereusnuntius/gowiki/internal/security"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type Security interface {
	GenerateKeyPair(ctx context.Context, actorID int64, actorURI string, keyType model.KeyType) error
	SavePublicKey(ctx context.Context, key *model.PublicKey) error
	PublicKeyExists(ctx context.Context, keyIRI string) (bool, error)
	PostSigned(ctx context.Context, inbox string, payload any, actorIRI string) error
}

type Actors struct {
	Config    config.WikiConfig
	Security  Security
	Store     Store
	TxManager *txdb.TxManager
	Publisher processor.Client
}

func New(config config.WikiConfig, store Store, security Security, manager *txdb.TxManager, publisher processor.Client) *Actors {
	return &Actors{
		Config:    config,
		Store:     store,
		TxManager: manager,
		Security:  security,
		Publisher: publisher,
	}
}

func (a *Actors) Initialize(ctx context.Context) error {
	actor, err := a.Store.GetActorByIRI(ctx, a.Config.URL.String())
	if err == nil || !wikierr.Is(err, wikierr.ErrNotFound) {
		return err
	}

	wikilog.Logger.Debug().Msg("saving instance actor on database")
	inbox := a.Config.URL.JoinPath("inbox").String()
	actor = model.Actor{
		URI:      a.Config.URL.String(),
		Type:     "Service",
		Username: a.Config.Name,
		Host:     a.Config.URL.Host,
		Inbox:    inbox,
		// SharedInbox: inbox,
		// SharedInboxID: ,
		Outbox:    a.Config.URL.JoinPath("outbox").String(),
		Followers: a.Config.URL.JoinPath("followers").String(),
		Following: a.Config.URL.JoinPath("following").String(),
		URL:       a.Config.URL.String(),
		Published: time.Now().UTC(),
	}

	return a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		wikilog.Logger.Debug().Msg("creating shared inbox for instance actor")
		id, err := a.Store.CreateSharedInboxIfNotExists(ctx, inbox)
		if err != nil {
			return err
		}

		actor.SharedInboxID = id
		if err = a.Store.SaveActor(ctx, &actor); err != nil {
			return err
		}

		if err = a.Security.GenerateKeyPair(ctx, actor.ID, actor.URI, model.RSAKey); err != nil {
			return err
		}
		wikilog.Logger.Debug().Str("actor", actor.URI).Msg("generated key pair for instance actor")

		return nil
	})
}

func (a *Actors) ActorExists(ctx context.Context, id string) (bool, error) {
	return a.Store.ActorExists(ctx, id)
}

func (a *Actors) GetActorIdForUser(ctx context.Context, userID int64) (int64, error) {
	return a.Store.GetActorIdForUser(ctx, userID)
}

func (a *Actors) CacheRemoteActor(ctx context.Context, actor activitystreams.Actor) (model.Actor, error) {
	iri, err := url.Parse(actor.Id)
	if err != nil {
		return model.Actor{}, err
	}

	actorInternal, err := a.GetActorByIRI(ctx, actor.Id)
	if err == nil || !wikierr.Is(err, wikierr.ErrNotFound) {
		return actorInternal, err
	}

	actorInternal = model.Actor{
		URI:         actor.Id,
		Type:        actor.Type,
		Username:    actor.Username,
		Host:        iri.Host,
		DisplayName: actor.Name,
		Summary:     actor.Summary,
		Inbox:       actor.Inbox,
		// SharedInbox: "", // TODO
		// SharedInboxID: ,
		Outbox:    actor.Outbox,
		Followers: actor.Followers,
		Following: actor.Following,
		// PublicKey: actor.PublicKey.,
		// URL: actor,
		Published: actor.Published,
		Updated:   actor.Updated,
	}

	err = a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		err := a.Store.SaveActor(ctx, &actorInternal)
		if err != nil {
			return err
		}

		key := actor.PublicKey
		exists, err := a.Security.PublicKeyExists(ctx, key.URI)
		if err != nil {
			return err
		}

		if !exists {
			key.OwnerID = actorInternal.ID
			return a.Security.SavePublicKey(ctx, &key)
		}

		return nil
	})
	if err != nil {
		return model.Actor{}, err
	}

	return actorInternal, nil
}

func (a *Actors) CreateLocalActor(ctx context.Context, username string, userID int64) error {
	// /u/{username}
	uri := a.Config.URL.JoinPath("u", username).String()
	actor := model.Actor{
		URI:       uri,
		Type:      "Person",
		Username:  username,
		Host:      a.Config.Host,
		Inbox:     uri + "/inbox",
		Outbox:    uri + "/outbox",
		Followers: uri + "/followers",
		Following: uri + "/following",
		URL:       uri,
		UserID:    userID,
		Published: time.Now().UTC(),
	}

	err := a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		id, err := a.Store.CreateSharedInboxIfNotExists(ctx, a.Config.SharedInbox)
		if err != nil {
			return err
		}

		actor.SharedInboxID = id
		err = a.Store.SaveActor(ctx, &actor)
		if err != nil {
			return err
		}

		if err = a.Security.GenerateKeyPair(ctx, actor.ID, actor.URI, model.RSAKey); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (a *Actors) GetLocalActor(ctx context.Context, username string) (model.Actor, error) {
	return a.Store.GetActorByHandle(ctx, username, a.Config.Host)
}

func (a *Actors) GetActorByID(ctx context.Context, id int64) (model.Actor, error) {
	return a.Store.GetActorByID(ctx, id)
}

func (a *Actors) GetActorByIRI(ctx context.Context, iri string) (model.Actor, error) {
	return a.Store.GetActorByIRI(ctx, iri)
}

func (a *Actors) Follow(ctx context.Context, follow *model.Follow) error {
	wikilog.Logger.Debug().
		Str("follower", follow.FollowerIRI).
		Str("followee", follow.FolloweeIRI).
		Msg("follow")
	actor, err := url.Parse(follow.FollowerIRI)
	if err != nil {
		return err
	}

	object, err := url.Parse(follow.FolloweeIRI)
	if err != nil {
		return err
	}

	followee, err := a.Store.GetActorByIRI(ctx, follow.FolloweeIRI)
	if err != nil {
		return err
	}

	follower, err := a.Store.GetActorByIRI(ctx, follow.FollowerIRI)
	if err != nil {
		return err
	}

	follow.FollowerID = follower.ID
	follow.FolloweeID = followee.ID

	if actor.Host == a.Config.Host {
		// Create ID for Follow activity
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		followIRI := a.Config.URL.JoinPath("follows", uuid.String()).String()
		follow.IRI = followIRI
		follow.Published = time.Now().UTC()

		return a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
			// A local actor (which probably only be the instance actor) wants to follow...
			switch object.Host {
			case a.Config.Host: // ... a local actor?
				return nil // Not implemented yet.
			default: // ... a remote actor.

				if err = a.Store.SaveFollow(ctx, follow); err != nil {
					return err
				}

				activity := streams.FollowAS(follow)
				return a.Security.PostSigned(ctx, followee.Inbox, activity, follow.FollowerIRI)
			}
		})
	} else if object.Host == a.Config.Host { // A foreign actor wants to follow a local actor.
		// TODO: For now accept immediately. Later implement manual review of follow requests.
		return a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
			accept := streams.Accept(follow.IRI, follow.FolloweeIRI)
			job := security.PostActivityJobArgs{
				Activity: accept,
				Inbox:    follower.Inbox,
				Actor:    followee.URI,
			}
			if err = a.Publisher.Publish(ctx, job); err != nil {
				return err
			}

			follow.Accepted = true
			if err = a.Store.SaveFollow(ctx, follow); err != nil {
				return err
			}
			wikilog.Logger.Debug().Msg("received and processed follow activity")
			return nil
		})
	}

	// A foreign actor wants to follow another foreign actor; we don't give a shit.
	return nil
}

func (a *Actors) AcceptFollow(ctx context.Context, actorIRI, followIRI string) error {
	url, err := url.Parse(followIRI)
	if err != nil {
		return err
	}

	if url.Host != a.Config.Host {
		return fmt.Errorf("activity %s did not come from this server", followIRI)
	}
	follow, err := a.Store.GetFollow(ctx, followIRI)
	if err != nil {
		return err
	}

	if follow.FolloweeIRI != actorIRI {
		return wikierr.New(wikierr.ErrForbidden, "wrong actor")
	}

	return a.Store.SetFollowAccepted(ctx, followIRI, true)
}

// TODO: add pagination.
func (a *Actors) GetFollowers(ctx context.Context, actor string) ([]string, error) {
	return a.Store.GetFollowers(ctx, actor)
}
