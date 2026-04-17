package articles

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/view"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type Handler struct {
	ArticleService *ArticleService
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /a/{slug}", handler.Read)
	mux.HandleFunc("GET /a/{slug}/edit", authhelpers.Authenticated(handler.ArticleEditor))
	mux.HandleFunc("POST /a/{slug}/edit", authhelpers.Authenticated(handler.Submit))

	mux.HandleFunc("GET /a/{host}/{slug}", handler.Read)

	mux.HandleFunc("POST /preview", authhelpers.Authenticated(handler.Preview))
	mux.HandleFunc("GET /search", handler.Search)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Search()")
		return
	}

	query := p.GetString("query")
	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") { // Replace this with a regex
		article, err := h.ArticleService.CacheRemoteArticle(p.Ctx, query)
		if err != nil {
			wikilog.Logger.Error().Err(err).Msg("h.ArticleService.fetchArticle()")
			p.HandleError(err)
			return
		}

		p.Redirect(
			fmt.Sprintf("/a/%s/%s", article.Article.Host, article.Article.Slug),
			"#content",
		)
		return
	}

	articles, err := h.ArticleService.SearchArticles(p.Ctx, query)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("h.ArticleService.SearchArticles()")
		p.HandleError(err)
		return
	}
	fmt.Println(articles)

	p.Content.Data = articles

	p.Render("search/results.html")
}

func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Read()")
		return
	}

	slug := r.PathValue("slug")
	host := r.PathValue("host")

	req := model.ArticleRequest{
		Slug: slug,
		Host: host,
	}

	content, err := h.ArticleService.ArticleContent(p.Ctx, &req)
	if err != nil {
		p.HandleError(err)
		return
	}

	view := view.Article{
		Slug:          content.Article.Slug,
		Host:          content.Article.Host,
		Content:       content.Content,
		ArticleHeader: articleHeader(slug),
	}

	p.Content.Data = view
	p.AddTemplate("articles/read.html")
	p.AddTemplate("articles/header.html")
	p.Render("articles/index.html")
}

func (h *Handler) ArticleEditor(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("ArticleEditor()")
		return
	}

	slug := r.PathValue("slug")

	req := model.ArticleRequest{
		Slug: slug,
	}
	content, err := h.ArticleService.ArticleContent(p.Ctx, &req)
	if err != nil && !errors.Is(err, wikierr.ErrNotFound) {
		wikilog.Logger.Error().Err(err).Msg("handler.ArticleService.ArticleContent()")
		return
	}

	view := view.Editor{
		Slug:          content.Article.Slug,
		Host:          content.Article.Host,
		Content:       content.Content,
		EditSummary:   "",
		LastModified:  content.Updated,
		ActionURL:     "/a/" + slug + "/edit",
		ArticleHeader: articleHeader(content.Article.Slug),
	}

	p.Content.Title = view.Slug
	p.Content.Data = view
	p.AddTemplate("articles/write.html")
	p.AddTemplate("articles/header.html")
	p.Render("articles/index.html")
}

func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Preview()")
	}

	content := p.GetString("content")
	p.Write(content)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Submit()")
		return
	}

	slug := r.PathValue("slug")

	content := p.GetString("content")
	summary := p.GetString("summary")

	edit := model.ArticleEdit{
		ActorID: p.Content.Session.User.ID,
		Slug:    slug,
		// Host: ,
		// IRI: ,
		// Lang: ,
		NewContent: content,
		Summary:    summary,
	}
	if err = h.ArticleService.LocalEdit(p.Ctx, edit); err != nil {
		wikilog.Logger.Error().Err(err).Msg("h.ArticleService.LocalEdit()")
		p.HandleError(err)
		return
	}

	h.Read(w, r)
}

func articleHeader(slug string) view.ArticleHeader {
	url := "/a/" + slug
	return view.ArticleHeader{
		ArticleURL: url,
		HistoryURL: url + "/history",
		EditURL:    url + "/edit",
	}
}
