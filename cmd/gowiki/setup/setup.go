package setup

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/auth"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/defaulthandler"
	"github.com/sidereusnuntius/gowiki/internal/federation/client"
	httphelpers "github.com/sidereusnuntius/gowiki/internal/helpers/http"
	"github.com/sidereusnuntius/gowiki/internal/search"
	"github.com/sidereusnuntius/gowiki/internal/tests"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type Wiki struct {
	Config         config.WikiConfig
	DB             *sql.DB
	TxManager      *txdb.TxManager
	AuthHandler    *auth.Handler
	DefaultHandler *defaulthandler.DefaultHandler
	Mux            *http.ServeMux
	Server         *http.Server
}

func SetupWiki(db *sql.DB, search *search.Search) Wiki {
	config := tests.TestConfig("http://localhost:8080")

	// Transaction manager.
	tm := &txdb.TxManager{
		DB: db,
	}

	client := client.New()

	// Setup data stores.
	authStore := setupAuthStore(db)
	sessionStore := setupSessionsStore(db)
	actorsStore := setupActorsStore(db)
	articlesStore := setupArticleStore(db)
	keyStore := setupKeyStore(db)

	// Setup services.
	actors := setupActorsService(actorsStore, keyStore, tm)
	auth := setupAuth(authStore, sessionStore, actors, tm)
	articles := setupArticles(config, articlesStore, tm, search, client)

	// Setup handlers.
	articlesHandler := setupArticlesHandler(articles)
	authHandler := setupAuthHandler(auth)
	defaultHandler := defaulthandler.New(articles)

	activityPubHandler := setupActivityPubHandler(actors)

	// Wire HTTP routing and handlers.
	mux := http.NewServeMux()
	defaultHandler.RegisterRoutes(mux)
	authHandler.RegisterRoutes(mux)
	articlesHandler.RegisterRoutes(mux)

	// Handler for serving static files such as CSS and Javascript.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Mux for handling Activitypub requests
	apMux := http.NewServeMux()
	activityPubHandler.RegisterRoutes(apMux)

	fedweb := httphelpers.FedWebMux(apMux, authHandler.SessionMiddleware(mux))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: fedweb,
	}

	return Wiki{
		Config:         config,
		DB:             db,
		TxManager:      tm,
		AuthHandler:    authHandler,
		DefaultHandler: defaultHandler,
		Mux:            mux,
		Server:         server,
	}
}
