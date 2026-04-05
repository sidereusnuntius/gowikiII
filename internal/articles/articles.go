package articles

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/search"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type ArticleService struct {
	Search    *search.Search
	Store     ArticleStore
	TxManager *txdb.TxManager
	Diffs     *diffmatchpatch.DiffMatchPatch
}

func New(store ArticleStore, manager *txdb.TxManager, search *search.Search) *ArticleService {
	return &ArticleService{
		Search:    search,
		Store:     store,
		TxManager: manager,
		Diffs:     diffmatchpatch.New(),
	}
}

func (as *ArticleService) SearchArticles(ctx context.Context, query string) ([]model.Article, error) {
	res, err := as.Search.SearchArticles(query)
	if err != nil {
		return nil, err
	}

	ids := make([]string, res.Hits.Len())
	for _, h := range res.Hits {
		fmt.Println(h.ID)
		ids = append(ids, h.ID)
	}
	fmt.Println(ids)

	results, err := as.Store.SearchArticles(ctx, ids)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (as *ArticleService) ArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error) {
	normalizeRequest(req)

	content, err := as.Store.GetArticleContent(ctx, req)
	return content, err
}

func (as *ArticleService) LocalEdit(ctx context.Context, in model.ArticleEdit) error {
	normalizeEdit(&in)
	// Validate

	var new bool
	err := as.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		article, err := as.Store.GetArticle(ctx, in.Slug, in.Host)
		if err != nil {
			if !errors.Is(err, wikierr.ErrNotFound) {
				return fmt.Errorf("as.Store.GetArticle(): %w", err)
			}

			article, err = as.CreateArticle(ctx, in.Slug, in.Host)
			if err != nil {
				return fmt.Errorf("as.CreateArticle(): %w", err)
			}
			wikilog.Logger.Debug().
				Str("slug", article.Slug).
				Str("host", article.Host).
				Msg("created new article")
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
			if err != nil && !errors.Is(err, wikierr.ErrNotFound) {
				return fmt.Errorf("as.Store.GetArticleContent(): %w", err)
			}
		} else {
			populateEmptyLocalizedArticle(&content, &article, &in)
		}

		err = as.EditArticleContent(ctx, &content, in)
		if err != nil {
			return fmt.Errorf("as.EditArticleContent(): %w", err)
		}

		return as.Search.IndexArticle(&content)
	})

	return err
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

	wikilog.Logger.Debug().
		Str("slug", article.Slug).
		Str("host", article.Host).
		Msg("saving article")
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
func (as *ArticleService) EditArticleContent(ctx context.Context, content *model.ArticleContent, edit model.ArticleEdit) error {
	var err error
	diff := as.Diffs.DiffMain(content.Content, edit.NewContent, false)
	delta := as.Diffs.DiffToDelta(diff)

	content.Content = edit.NewContent

	if content.ID != 0 {
		err = as.Store.UpdateLocalizedArticle(ctx, content)
	} else {
		err = as.Store.SaveArticleContent(ctx, content)
	}

	if err != nil {
		return err
	}

	buf := make([]byte, 24)
	rand.Read(buf)
	code := base64.URLEncoding.EncodeToString(buf)

	iri, _ := url.Parse(content.Article.IRI)
	iri = iri.JoinPath("edits", code)

	revision := model.Revision{
		Code:    code,
		IRI:     iri.String(),
		Diff:    delta,
		Summary: edit.Summary,
		// Prev: 0,
		ArticleID: content.Article.ID,
		Published: time.Now().UTC(),
		ActorID:   edit.ActorID,
		ActorIRI:  edit.ActorIRI,
	}

	return as.Store.SaveRevision(ctx, &revision)
}

// Homepage finds the homepage of the given wiki, if it exists. We expect each wiki to have a home article,
// which contains the text to be displayed in the home page. If wiki is empty, then the homepage of the local
// wiki is returned.
func (as *ArticleService) Homepage(ctx context.Context, wiki, locale string) (model.ArticleContent, error) {
	if len(wiki) == 0 {
		wiki = config.Config.Host
	}

	req := model.ArticleRequest{
		Slug: "home",
		Host: wiki,
	}
	return as.ArticleContent(ctx, &req)
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

func normalizeRequest(req *model.ArticleRequest) {
	req.Slug = strings.ToLower(
		strings.TrimSpace(req.Slug),
	)

	if len(req.Host) == 0 {
		req.Host = config.Config.Host
	} else {
		req.Host = strings.ToLower(
			strings.TrimSpace(req.Host),
		)
	}
}

func populateEmptyLocalizedArticle(content *model.ArticleContent, article *model.Article, edit *model.ArticleEdit) {
	*content = model.ArticleContent{
		Article: *article,
		// Lang: ,
		// Title: ,
		URL:       "/a/" + article.Slug,
		Published: time.Now().UTC(),
	}
}
