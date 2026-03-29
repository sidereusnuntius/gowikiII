package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/articles"
	articlesql "github.com/sidereusnuntius/gowiki/internal/articles/sql"
	"github.com/sidereusnuntius/gowiki/internal/search"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

func setupArticleStore(db *sql.DB) *articlesql.ArticleStore {
	store := articlesql.New(db)
	return store
}

func setupArticles(store articles.ArticleStore, tm *txdb.TxManager, search *search.Search) *articles.ArticleService {
	return articles.New(store, tm, search)
}

func setupArticlesHandler(service *articles.ArticleService) *articles.Handler {
	return &articles.Handler{
		ArticleService: service,
	}
}
