package actors

import (
	"context"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type KeyStore interface {
	GenerateKeyPair(ctx context.Context, actorID int64, actorURI string, keyType model.KeyType) error
}

type Actors struct {
	Keys      KeyStore
	Store     Store
	TxManager *txdb.TxManager
}

func New(store Store, keys KeyStore, manager *txdb.TxManager) Actors {
	return Actors{
		Store:     store,
		TxManager: manager,
		Keys:      keys,
	}
}

func (a *Actors) CreateLocalActor(ctx context.Context, username string, userID int64) error {
	// /u/{username}
	uri := config.Config.URL.JoinPath("u", username).String()
	actor := model.Actor{
		URI:       uri,
		Type:      "Person",
		Username:  username,
		Host:      config.Config.Host,
		Inbox:     uri + "/inbox",
		Outbox:    uri + "/outbox",
		Followers: uri + "/followers",
		Following: uri + "/following",
		URL:       uri,
		UserID:    userID,
		Published: time.Now().UTC(),
	}

	err := a.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		id, err := a.Store.CreateSharedInboxIfNotExists(ctx, config.Config.SharedInbox)
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
