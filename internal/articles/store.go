package articles

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type ArticleStore interface {
	ArticleExistsLocally(ctx context.Context, slug, host string) (bool, error)
	SaveArticle(ctx context.Context, article *model.Article) error
	GetArticle(ctx context.Context, slug, host string) (model.Article, error)

	GetArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error)
	SaveArticleContent(ctx context.Context, content *model.ArticleContent) error
	// UpdateArticleContent updates the content of a localized article content.
	UpdateLocalizedArticle(ctx context.Context, content *model.ArticleContent) error

	SaveRevision(ctx context.Context, revision *model.Revision) error
}
