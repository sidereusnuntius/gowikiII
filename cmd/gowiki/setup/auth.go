package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/auth"
	authstore "github.com/sidereusnuntius/gowiki/internal/auth/sql"
	"github.com/sidereusnuntius/gowiki/internal/config"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

func setupAuthStore(db *sql.DB) *authstore.AuthStore {
	store := authstore.New(db)
	return &store
}

func setupAuth(authStore *authstore.AuthStore, sessionStore auth.SessionStore, actorsService auth.ActorService, tm *txdb.TxManager) *auth.Auth {
	return &auth.Auth{
		TxManager: tm,
		Actors:    actorsService,
		Store:     authStore,
		Sessions:  sessionStore,
	}
}

func setupAuthHandler(config *config.WikiConfig, service *auth.Auth) *auth.Handler {
	return &auth.Handler{
		Config:       config,
		AuthService:  service,
		SessionStore: service.Sessions,
	}
}
