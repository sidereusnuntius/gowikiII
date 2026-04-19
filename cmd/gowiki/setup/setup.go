package setup

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/actors"
	"github.com/sidereusnuntius/gowiki/internal/articles"
	"github.com/sidereusnuntius/gowiki/internal/auth"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/defaulthandler"
	"github.com/sidereusnuntius/gowiki/internal/federation"
	"github.com/sidereusnuntius/gowiki/internal/federation/client"
	httphelpers "github.com/sidereusnuntius/gowiki/internal/helpers/http"
	"github.com/sidereusnuntius/gowiki/internal/search"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type Wiki struct {
	Config          config.WikiConfig
	DB              *sql.DB
	TxManager       *txdb.TxManager
	Actors          *actors.Actors
	AuthHandler     *auth.Handler
	ArticlesHandler *articles.Handler
	DefaultHandler  *defaulthandler.DefaultHandler
	Mux             *http.ServeMux
	Server          *http.Server
}

func SetupWiki(config config.WikiConfig, db *sql.DB, search *search.Search) (Wiki, error) {
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
	security, err := setupSecurity(config, db, client)
	if err != nil {
		return Wiki{}, err
	}
	hostsStore := setupHostsStore(db)

	// Setup services.
	actors := setupActorsService(config, actorsStore, security, tm)
	auth := setupAuth(authStore, sessionStore, actors, tm)
	articles := setupArticles(config, articlesStore, tm, search, client, security, actors)
	federation := federation.New(config, hostsStore, client, actors, articles, nil)

	// Setup handlers.
	articlesHandler := setupArticlesHandler(articles)
	authHandler := setupAuthHandler(auth)
	defaultHandler := defaulthandler.New(articles)

	activityPubHandler := setupActivityPubHandler(federation, articles, actors, security)

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
		Config:          config,
		DB:              db,
		TxManager:       tm,
		Actors:          actors,
		AuthHandler:     authHandler,
		ArticlesHandler: articlesHandler,
		DefaultHandler:  defaultHandler,
		Mux:             mux,
		Server:          server,
	}, nil
}
