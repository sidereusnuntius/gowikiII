package articles

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type ArticleService struct {
	Store     Archive
	TxManager *txdb.TxManager
	Diffs     *diffmatchpatch.DiffMatchPatch
}

func New(store Archive, manager *txdb.TxManager) *ArticleService {
	return &ArticleService{
		Store:     store,
		TxManager: manager,
	}
}

func (as *ArticleService) LocalEdit(ctx context.Context, userID string, in model.ArticleEdit) error {
	normalizeEdit(&in)
	// Validate

	var new bool
	err := as.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		article, err := as.Store.GetArticle(ctx, in.Slug, in.Host)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			article, err = as.CreateArticle(ctx, in.Slug, in.Host)
			if err != nil {
				return err
			}
			new = true
		}

		var content model.ArticleContent
		if !new {
			req := model.ArticleRequest{
				Slug: article.Slug,
				Host: article.Host,
				// IRI   string
				// Langs []string
			}
			content, err = as.Store.GetArticleContent(ctx, &req)
			// There is already a localized version of this article in the provided language,
			// so we update it.
			if err == nil {
				return as.UpdateArticle(ctx, content, in)
			}

			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		return nil
	})

}

// CreateArticle creates an article object, generating a new IRI for it, and saves it in the data store.
func (as *ArticleService) CreateArticle(ctx context.Context, slug, host string) (model.Article, error) {
	var iri string
	if len(host) == 0 || host == config.Config.Host {
		iri = config.Config.URL.JoinPath("a", slug).String()
	} else {
		iri = config.Config.URL.JoinPath("a", slug+"@"+host).String()
	}

	published := time.Now().UTC()
	article := model.Article{
		Slug:      slug,
		Host:      host,
		IRI:       iri,
		Published: published,
		// FederatedEdits: ,
		// Restricted: ,
	}

	err := as.Store.SaveArticle(ctx, &article)
	if err != nil {
		return model.Article{}, err
	}
	return article, nil
}

func (as *ArticleService) UpdateArticle(ctx context.Context, content model.ArticleContent, edit model.ArticleEdit) error {

	return nil
}

// EditArticleContent handles creation and modification of localized article instances. If the provided ArticleContent is not the zero value,
// it is assumed to be an existing localized article, and so this method will apply the edit to the localized article and persist it.
// Otherwise, it creates the localized article. In both cases it creates a new revision.
func (as *ArticleService) EditArticleContent(ctx context.Context, content model.ArticleContent, edit model.ArticleEdit) error {
	diff := as.Diffs.DiffMain(content.Content, edit.NewContent, false)
	delta := as.Diffs.DiffToDelta(diff)
	
	if content.ID != 0 {
		
	}
}

func (as *ArticleService) LocalCreate(ctx context.Context, in model.ArticleEdit) error {
	published := time.Now().UTC()
	iri := config.Config.URL.JoinPath("a", in.Slug)

	article := model.Article{
		Slug:      in.Slug,
		Host:      config.Config.Host,
		IRI:       iri.String(),
		Published: published,
	}

	err := as.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		err := as.Store.SaveArticle(ctx, &article)
		if err != nil {
			return err
		}

		// content := model.ArticleContent{
		// 		Article: article,
		// 		Lang: in.Lang,
		// 		// Title: ,
		// 		Content: in.NewContent,
		// 		// Summary: ,
		// 		URL: ,
		// 		Published: ,
		// 		Updated: ,
		// 		Fetched: ,
		// }
	})
}

func normalizeEdit(in *model.ArticleEdit) {
	in.Slug = strings.ToLower(
		strings.TrimSpace(in.Slug),
	)

	if len(in.Host) == 0 {
		in.Host = config.Config.Host
	} else {
		in.Host = strings.ToLower(
			strings.TrimSpace(in.Host),
		)
	}
}
