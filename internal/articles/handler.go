package articles

import (
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

const (
	readTab = iota
	historyTab
	editTab
)

type Handler struct {
	ArticleService *ArticleService
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /a/{slug}", handler.Read)
	mux.HandleFunc("GET /a/{host}/{slug}", handler.Read)

	mux.HandleFunc("GET /a/{slug}/edit", authhelpers.Authenticated(handler.ArticleEditor))
	mux.HandleFunc("GET /a/{host}/{slug}/edit", authhelpers.Authenticated(handler.ArticleEditor))

	mux.HandleFunc("POST /a/{slug}/edit", authhelpers.Authenticated(handler.Submit))
	mux.HandleFunc("POST /a/{host}/{slug}/edit", authhelpers.Authenticated(handler.Submit))

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
	nextPage := p.GetInt64("next")

	filter := model.ArticleFilter{
		Query:    query,
		NextPage: nextPage,
	}
	result, err := h.ArticleService.SearchArticles(p.Ctx, filter)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("h.ArticleService.SearchArticles()")
		p.HandleError(err)
		return
	}

	p.Content.Data = result

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
		Slug:     content.Article.Slug,
		Host:     content.Article.Host,
		Content:  content.Content,
		Controls: articleControls(content.Article.Slug, readTab),
	}

	p.Content.Data = view
	p.AddTemplate("articles/read.html")
	p.AddTemplate("tabs.html")
	p.Render("articles/index.html")
}

func (h *Handler) ArticleEditor(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("ArticleEditor()")
		return
	}

	slug := r.PathValue("slug")
	host := r.PathValue("host")

	req := model.ArticleRequest{
		Slug: slug,
		Host: host,
	}
	content, err := h.ArticleService.ArticleContent(p.Ctx, &req)
	if err != nil && !wikierr.Is(err, wikierr.ErrNotFound) {
		wikilog.Logger.Error().Err(err).Msg("handler.ArticleService.ArticleContent()")
		return
	}

	view := view.Editor{
		Slug:         content.Article.Slug,
		Host:         content.Article.Host,
		Content:      content.Content,
		EditSummary:  "",
		LastModified: content.Updated,
		ActionURL:    "/a/" + slug + "/edit",
		Controls:     articleControls(content.Article.Slug, editTab),
	}

	p.Content.Title = view.Slug
	p.Content.Data = view
	p.AddTemplate("articles/write.html")
	p.AddTemplate("tabs.html")
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
	host := r.PathValue("host")

	content := p.GetString("content")
	summary := p.GetString("summary")

	edit := model.ArticleEdit{
		ActorID: p.Content.Session.User.ID,
		Slug:    slug,
		Host:    host,
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

func articleControls(slug string, currentTab int) [3]view.PageControl {
	url := "/a/" + slug
	tabs := [3]view.PageControl{
		{URL: url, Label: "Read"},
		{URL: url + "/history", Label: "History"},
		{URL: url + "/edit", Label: "Edit"},
	}

	tabs[currentTab].Selected = true
	return tabs
}
