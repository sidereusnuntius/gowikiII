package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/actors"
	actorstore "github.com/sidereusnuntius/gowiki/internal/actors/sql"
	"github.com/sidereusnuntius/gowiki/internal/config"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

func setupActorsStore(db *sql.DB) *actorstore.ActorsStore {
	return &actorstore.ActorsStore{
		DB: db,
	}
}

func setupActorsService(config config.WikiConfig, store actors.Store, keyStore actors.KeyStore, tm *txdb.TxManager) *actors.Actors {
	return &actors.Actors{
		Keys:      keyStore,
		Store:     store,
		TxManager: tm,
		Config:    config,
	}
}
