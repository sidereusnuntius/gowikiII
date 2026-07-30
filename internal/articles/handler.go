package articles

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
	"github.com/sidereusnuntius/gowiki/internal/i18n"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/view"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

const (
	ReadTab = iota
	HistoryTab
	EditTab
)

type Auth interface {
	GetActorIdForUser(ctx context.Context, userID int64) (int64, error)
}

type Handler struct {
	ArticleService *ArticleService
	AuthService    Auth
}

func NewHandler(articleService *ArticleService, auth Auth) Handler {
	return Handler{
		ArticleService: articleService,
		AuthService:    auth,
	}
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /a/{slug}", handler.Read)
	mux.HandleFunc("GET /a/{host}/{slug}", handler.Read)

	mux.HandleFunc("GET /recent", handler.RecentChanges)

	mux.HandleFunc("GET /a/{slug}/history", handler.ArticleHistory)
	mux.HandleFunc("GET /a/{host}/{slug}/history", handler.ArticleHistory)

	mux.HandleFunc("GET /revision/{id}", handler.Revision)
	mux.HandleFunc("POST /revision/{id}/undo", authhelpers.Authenticated(handler.RevisionUndo))

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

	p.Content.Title = fmt.Sprintf("%s: \"%s\"", i18n.T("SearchResultsFor"), query)
	p.Content.Data = result

	p.AddTemplate("tabs.html")
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
		if wikierr.Is(err, wikierr.ErrNotFound) {
			h.ArticleEditor(w, r)
			return
		}
		p.HandleError(err)
		return
	}

	html, err := h.ArticleService.Render(content.Content)
	if err != nil {
		p.HandleError(err)
		return
	}

	view := view.Article{
		Slug:    content.Article.Slug,
		Host:    content.Article.Host,
		Content: template.HTML(html),
	}

	p.Content.Title = content.Article.Slug
	p.Content.Controls = ArticleControls(&h.ArticleService.Config, content.Article.Slug, content.Article.Host, ReadTab)
	p.Content.Data = view
	p.AddTemplate("articles/read.html")
	p.AddTemplate("tabs.html")
	p.Render("articles/index.html")
}

func (h *Handler) ArticleHistory(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Read()")
		return
	}

	// TODO: get language.
	slug := r.PathValue("slug")
	host := r.PathValue("host")
	var after time.Time
	if query := r.URL.Query(); query.Has("after") {
		after, _ = time.Parse(time.RFC3339, query.Get("after"))
	}
	history, err := h.ArticleService.ArticleHistory(p.Ctx, slug, host, "", after)
	if err != nil {
		p.HandleError(err)
		return
	}

	view := view.HistoryView(&h.ArticleService.Config, history, false)
	p.Content.Title = slug
	p.Content.Controls = ArticleControls(&h.ArticleService.Config, slug, host, HistoryTab)
	p.Content.Data = view
	p.AddTemplate("tabs.html", "articles/history.html")
	p.Render("articles/index.html")
}

func prettyHtml(diffs []diffmatchpatch.Diff) template.HTML {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := strings.ReplaceAll(html.EscapeString(diff.Text), "\n", "<br>")
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("<ins>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</ins>")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("<del>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</del>")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("<span>")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("</span>")
		}
	}
	return template.HTML(buff.String())
}

func (h *Handler) RecentChanges(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("RecentChanges()")
		return
	}

	revisions, err := h.ArticleService.RecentChanges(p.Ctx, time.Now())
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("RecentChanges()")
		p.RenderError(err)
		return
	}

	view := view.HistoryView(&h.ArticleService.Config, revisions, true)

	p.Content.Data = view
	p.Content.Title = "Recent changes"
	p.AddTemplate("tabs.html")
	p.Render("recent.html")
}

func (h *Handler) Revision(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Read()")
		return
	}

	revisionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("Read()")
		return
	}
	revision, text, diffs, err := h.ArticleService.RevisionDiffs(p.Ctx, revisionID)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("h.ArticleService.RevisionDiffs")
		p.RenderError(err)
		return
	}

	fmt.Println(text)
	v := view.RevisionView{
		DiffHTML: prettyHtml(diffs),
		Summary:  revision.Summary,
		ActorURL: view.ActorURL(&h.ArticleService.Config, revision.ActorUsername, revision.ActorHost),
		Actor:    revision.ActorUsername,
		URL:      view.RevisionURL(revision.ID),
	}

	p.Content.Title = i18n.T("DiffTitle")
	p.Content.Controls = ArticleControls(&h.ArticleService.Config, revision.ArticleSlug, revision.ArticleHost, HistoryTab)
	p.Content.ExtraControls = []render.PageControl{
		{
			URL:    view.RevisionURL(revision.ID) + "/undo",
			Label:  "Undo",
			Method: http.MethodPost,
		},
	}
	p.Content.Data = v
	p.AddTemplate("tabs.html", "articles/revision.html")
	p.Render("articles/index.html")
}

func (h *Handler) RevisionUndo(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("RevisionUndo()")
		return
	}

	revisionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		p.NotFound()
		return
	}

	revision, text, _, err := h.ArticleService.RevisionDiffs(p.Ctx, revisionID)
	if err != nil {
		p.RenderError(err)
		return
	}

	v := view.Editor{
		Slug:        revision.ArticleSlug,
		Host:        revision.ArticleHost,
		Content:     text,
		Preview:     text,
		EditSummary: fmt.Sprintf("Undid revision %d by %s", revision.ID, revision.ActorUsername),
		ActionURL: view.ArticleEditURL(
			&h.ArticleService.Config,
			revision.ArticleSlug,
			revision.ArticleHost,
		),
	}

	p.Content.Title = view.ArticleTitle(&h.ArticleService.Config, revision.ArticleSlug, revision.ArticleHost)
	p.Content.Controls = ArticleControls(&h.ArticleService.Config, revision.ArticleSlug, revision.ArticleHost, EditTab)
	p.Content.Data = v
	p.AddTemplate("articles/write.html")
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
		Slug:         slug,
		Host:         host,
		Content:      content.Content,
		EditSummary:  "",
		LastModified: content.Updated,
		ActionURL: view.ArticleEditURL(
			&h.ArticleService.Config,
			slug,
			host,
		),
	}

	p.Content.Title = view.Slug
	p.Content.Controls = ArticleControls(&h.ArticleService.Config, content.Article.Slug, content.Article.Host, EditTab)
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

	source := p.GetString("content")
	content, err := h.ArticleService.Render(source)
	if err != nil {
		p.HandleError(err)
		return
	}

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

	actorID, err := h.AuthService.GetActorIdForUser(p.Ctx, p.Content.Session.User.ID)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("h.AuthService.GetActorIdForUser")
		p.RenderError(err)
		return
	}

	edit := model.ArticleEdit{
		ActorID: actorID,
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

	p.Redirect(view.ArticleURL(&h.ArticleService.Config, slug, host), "#content")
	h.Read(w, r)
}

func ArticleControls(config *config.WikiConfig, slug, host string, currentTab int) []render.PageControl {
	tabs := []render.PageControl{
		{URL: view.ArticleURL(config, slug, host), Label: "Read"},
		{URL: view.ArticleHistoryURL(config, slug, host), Label: "History"},
		{URL: view.ArticleEditURL(config, slug, host), Label: "Edit"},
	}

	tabs[currentTab].Selected = true
	return tabs
}
