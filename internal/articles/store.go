package articles

import (
	"context"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type ArticleStore interface {
	ArticleSlugExists(ctx context.Context, slug string) (bool, error)
	ArticleExistsLocally(ctx context.Context, iri string) (bool, error)
	SaveArticle(ctx context.Context, article *model.Article) error
	GetArticle(ctx context.Context, slug, host string) (model.Article, error)
	SearchArticles(ctx context.Context, iris []string, filter model.ArticleFilter) ([]model.Article, error)
	GetLocalizedArticleID(ctx context.Context, slug, host, lang string) (int64, error)

	GetArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error)
	SaveArticleContent(ctx context.Context, content *model.ArticleContent) error
	// UpdateArticleContent updates the content of a localized article content.
	UpdateLocalizedArticle(ctx context.Context, content *model.ArticleContent) error

	SaveRevision(ctx context.Context, revision *model.Revision) error
	RevisionHistory(ctx context.Context, localizedArticleID int64, after time.Time, limit int) ([]model.Revision, error)
	ArticleReverseHistory(ctx context.Context, localizedArticleID, targetRevisionID int64) ([]string, error)
	RevisionByID(ctx context.Context, revisionID int64) (model.Revision, error)
	RecentChanges(ctx context.Context, after time.Time, limit int) ([]model.Revision, error)
}
