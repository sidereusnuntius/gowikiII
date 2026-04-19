package actors

import (
	"context"
	"net/url"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type KeyStore interface {
	GenerateKeyPair(ctx context.Context, actorID int64, actorURI string, keyType model.KeyType) error
	SavePublicKey(ctx context.Context, key *model.PublicKey) error
	PublicKeyExists(ctx context.Context, keyIRI string) (bool, error)
}

type Actors struct {
	Config    config.WikiConfig
	Keys      KeyStore
	Store     Store
	TxManager *txdb.TxManager
}

func New(config config.WikiConfig, store Store, keys KeyStore, manager *txdb.TxManager) Actors {
	return Actors{
		Config:    config,
		Store:     store,
		TxManager: manager,
		Keys:      keys,
	}
}

func (a *Actors) ActorExists(ctx context.Context, id string) (bool, error) {
	return a.Store.ActorExists(ctx, id)
}

func (a *Actors) CacheRemoteActor(ctx context.Context, actor activitystreams.Actor) error {
	iri, err := url.Parse(actor.Id)
	if err != nil {
		return err
	}

	actorInternal := model.Actor{
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

	return a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		err := a.Store.SaveActor(ctx, &actorInternal)
		if err != nil {
			return err
		}

		key := actor.PublicKey
		exists, err := a.Keys.PublicKeyExists(ctx, key.URI)
		if err != nil {
			return err
		}

		if !exists {
			key.OwnerID = actorInternal.ID
			return a.Keys.SavePublicKey(ctx, &key)
		}

		return nil
	})
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

		if err = a.Keys.GenerateKeyPair(ctx, actor.ID, actor.URI, model.RSAKey); err != nil {
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
