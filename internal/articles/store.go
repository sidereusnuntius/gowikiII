package articles

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

// Archive is an article store, but with a fancy name.
type Archive interface {
	SaveArticle(ctx context.Context, article *model.Article) error

	GetArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error)
	SaveArticleContent(ctx context.Context, content *model.ArticleContent) error

	SaveRevision(ctx context.Context, revision *model.Revision) error
}
