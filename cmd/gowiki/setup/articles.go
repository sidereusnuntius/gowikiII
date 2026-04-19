package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/actors"
	"github.com/sidereusnuntius/gowiki/internal/articles"
	articlesql "github.com/sidereusnuntius/gowiki/internal/articles/sql"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/federation/client"
	"github.com/sidereusnuntius/gowiki/internal/search"
	"github.com/sidereusnuntius/gowiki/internal/security"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

func setupArticleStore(db *sql.DB) *articlesql.ArticleStore {
	store := articlesql.New(db)
	return store
}

func setupArticles(config config.WikiConfig, store articles.ArticleStore, tm *txdb.TxManager, search *search.Search, client client.Client, security *security.Security, actors *actors.Actors) *articles.ArticleService {
	return articles.New(config, store, tm, search, client, security, actors)
}

func setupArticlesHandler(service *articles.ArticleService) *articles.Handler {
	return &articles.Handler{
		ArticleService: service,
	}
}
