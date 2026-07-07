package defaulthandler

import (
	"context"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/articles"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/view"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type homepageFinder interface {
	Homepage(ctx context.Context, wiki, locale string) (model.ArticleContent, error)
}

type DefaultHandler struct {
	homepage homepageFinder
}

func New(articleService *articles.ArticleService) *DefaultHandler {
	return &DefaultHandler{
		homepage: articleService,
	}
}

func (h *DefaultHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.Home)
}

func (h *DefaultHandler) Home(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("DefaultHandler.Home()")
		return
	}

	content, err := h.homepage.Homepage(p.Ctx, "", "")
	if err != nil {
		p.HandleError(err)
		return
	}

	view := view.Article{
		Slug:    "home",
		Content: content.Content,
	}

	p.Content.Title = "home"
	p.Content.Controls = articles.ArticleControls("home", articles.ReadTab)
	p.Content.Data = view
	p.AddTemplate("articles/read.html")
	p.AddTemplate("tabs.html")
	p.Render("articles/index.html")
}
