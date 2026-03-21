package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/auth"
	authstore "github.com/sidereusnuntius/gowiki/internal/auth/sql"
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

func setupAuthHandler(service *auth.Auth) *auth.Handler {
	return &auth.Handler{
		AuthService:  service,
		SessionStore: service.Sessions,
	}
}
