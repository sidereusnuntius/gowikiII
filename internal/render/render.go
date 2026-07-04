package render

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
	"github.com/sidereusnuntius/gowiki/templates"
)

var templatesHome string

type CommonView struct {
	Title         string
	Session       *model.Session
	Authenticated bool
	Data          any
}

type Page struct {
	template     *template.Template
	subtemplates []string
	writer       http.ResponseWriter
	req          *http.Request
	status       int
	// Indicates whether the request was made by Fixi.js. If true, the page can be updated partially.
	isFx    bool
	Ctx     context.Context
	Content CommonView
}

func Init(w http.ResponseWriter, r *http.Request) (*Page, error) {
	var err error
	page := Page{
		writer: w,
		req:    r,
		isFx:   r.Header.Get("FX-Request") == "true",
		Ctx:    r.Context(),
	}

	session, ok := authhelpers.GetSession(page.Ctx)
	if ok {
		page.Content.Session = session
		page.Content.Authenticated = true
	}

	if !page.isFx {
		// If it's not a fixi request, then we need to render the whole page.
		page.template, err = template.ParseFS(
			templates.TemplatesFS,
			"index.html",
			"container.html",
			"header.html",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read common layout from file: %w", err)
		}
	}

	return &page, nil
}

func (p *Page) GetString(name string) string {
	return p.req.FormValue(name)
}

func (p *Page) GetInt64(name string) int64 {
	n := p.req.FormValue(name)
	if len(n) == 0 {
		return 0
	}

	n64, _ := strconv.ParseInt(n, 10, 64)
	return n64
}

func (p *Page) AddTemplate(paths ...string) {
	// for i := range paths {
	// 	paths[i] = paths[i]
	// }
	p.subtemplates = append(p.subtemplates, paths...)
}

func (p *Page) Write(text string) error {
	_, err := p.writer.Write([]byte(text))
	return err
}

func (p *Page) RenderError(err error) error {
	p.writer.WriteHeader(http.StatusInternalServerError) // TODO: better handle these errors.
	message := []byte(err.Error())
	_, err2 := p.writer.Write(message)
	return err2
}

func (p *Page) PatchElement(path, templateName, selector string) {
	wikilog.Logger.Debug().Str("template", path).Str("selector", selector).Msg("patching element")
	p.writer.Header().Set("FX-target", selector)
	p.writer.WriteHeader(p.status)

	tmpl, err := template.ParseFS(templates.TemplatesFS, path)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to parse template")
	}

	if err = tmpl.ExecuteTemplate(p.writer, templateName, p.Content); err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("template", path).
			Msg("failed to execute template")
	}
}

func (p *Page) ReloadPage(path string) {
	var err error
	wikilog.Logger.Debug().Str("path", path).Msg("reloading full page")
	header := p.writer.Header()
	header.Add("Fx-Target", "#page-container")
	header.Add("Content-Type", "text/html")
	tmpl, err := template.ParseFS(
		templates.TemplatesFS,
		"container.html",
		"header.html",
		path,
	)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to parse template")
	}

	if err = tmpl.ExecuteTemplate(p.writer, "container", p.Content); err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to reload page")
	}
}

func (p *Page) SetHeader(header int) {
	p.writer.WriteHeader(header)
}

func (p *Page) Redirect(url, target string) {
	if !p.isFx {
		http.Redirect(p.writer, p.req, url, http.StatusSeeOther)
		return
	}

	header := p.writer.Header()
	header.Add("Fx-Redirect", url)
	header.Add("Fx-Redirect-Target", target)
	p.status = http.StatusSeeOther
}

func (p *Page) Render(path string) {
	l := log.Debug().Str("template path", path)
	if p.isFx {
		header := p.writer.Header()
		//header.Add("FX-target", "#content")
		if len(p.Content.Title) > 0 {
			header.Set("Set-Title", p.Content.Title)
			l.Str("page title", p.Content.Title)
		}

		p.subtemplates = append(p.subtemplates, path)

		tmpl, err := template.ParseFS(templates.TemplatesFS, p.subtemplates...)
		if err != nil {
			wikilog.Logger.Error().Err(err).Str("template", path).Msg("failed to parse template file")
		}

		err = tmpl.ExecuteTemplate(p.writer, "content", p.Content)
		if err != nil {
			wikilog.Logger.Error().Err(err).Str("template", path).Msg("failed to execute template")
		}
		l.Msg("patched element via Fixi")

		return
	}
	l.Msg("doing full page render")
	p.render(path)
}

func (p *Page) render(path string) {
	var err error
	p.subtemplates = append(p.subtemplates, path)

	p.template, err = p.template.ParseFS(templates.TemplatesFS, p.subtemplates...)
	if err != nil {
		wikilog.Logger.Error().Err(err).Strs("subtemplates", p.subtemplates).Msg("failed to read subtemplates from files")
	}

	if err = p.template.Execute(p.writer, p.Content); err != nil {
		wikilog.Logger.Error().Err(err).Str("template", path).Msg("failed to execute template")
	}
	wikilog.Logger.Debug().Str("template", path).Strs("subtemplates", p.subtemplates).Msg("rendered template")
}
