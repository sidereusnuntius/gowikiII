package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/actors"
	actorstore "github.com/sidereusnuntius/gowiki/internal/actors/sql"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

func setupActorsStore(db *sql.DB) *actorstore.ActorsStore {
	return &actorstore.ActorsStore{
		DB: db,
	}
}

func setupActorsService(config config.WikiConfig, store actors.Store, security actors.Security, tm *txdb.TxManager, client processor.Client) *actors.Actors {
	return actors.New(config, store, security, tm, client)
}
