package setup

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/auth"
	"github.com/sidereusnuntius/gowiki/internal/config"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type Wiki struct {
	DB          *sql.DB
	TxManager   *txdb.TxManager
	AuthHandler *auth.Handler
	Mux         *http.ServeMux
	Server      *http.Server
}

func SetupWiki(db *sql.DB) Wiki {
	// Transaction manager.
	tm := &txdb.TxManager{
		DB: db,
	}

	// Setup data stores.
	authStore := setupAuthStore(db)
	sessionStore := setupSessionsStore(db)
	actorsStore := setupActorsStore(db)
	keyStore := setupKeyStore(db)

	// Setup services.
	actors := setupActorsService(actorsStore, keyStore, tm)
	auth := setupAuth(authStore, sessionStore, actors, tm)

	// Setup handlers.
	authHandler := setupAuthHandler(auth)

	// Wire HTTP routing and handlers.
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)

	// Handler for serving static files such as CSS and Javascript.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Port),
		Handler: authHandler.SessionMiddleware(mux),
	}

	return Wiki{
		DB:          db,
		TxManager:   tm,
		AuthHandler: authHandler,
		Mux:         mux,
		Server:      server,
	}
}
