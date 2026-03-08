package render

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/starfederation/datastar-go/datastar"
)

var templatesHome string

type Page struct {
	sse          *datastar.ServerSentEventGenerator
	signals      map[string]any
	template     *template.Template
	subtemplates []string
	writer       http.ResponseWriter
	req          *http.Request
	isDatastar   bool
	Title        string
	Ctx          context.Context
	Data         any
}

func Init(w http.ResponseWriter, r *http.Request) (*Page, error) {
	var err error
	page := Page{
		writer:     w,
		req:        r,
		isDatastar: r.Header.Get("Datastar-Request") == "true",
		Ctx:        r.Context(),
	}

	if page.isDatastar {
		fmt.Println("it's a datastar request!")
		err = datastar.ReadSignals(r, &page.signals)
		if err != nil {
			return nil, fmt.Errorf("failed to read signals from request") // Improve errors.
		}
		page.sse = datastar.NewSSE(w, r)
	} else {
		page.template, err = template.ParseFiles(TemplatePath("index.html"))
		if err != nil {
			return nil, fmt.Errorf("failed to read common layout from file: %w", err)
		}
		if err = r.ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse request body: %w", err)
		}
	}

	return &page, nil
}

func (p *Page) GetString(name string) string {
	if p.isDatastar {
		return p.signals[name].(string)
	}
	return p.req.Form.Get(name)
}

func (p *Page) AddTemplate(path string) {
	p.subtemplates = append(p.subtemplates, TemplatePath(path))
}

func (p *Page) Render(template string) error {
	var err error
	p.subtemplates = append(p.subtemplates, TemplatePath(template))

	p.template, err = p.template.ParseFiles(p.subtemplates...)
	if err != nil {
		return fmt.Errorf("failed to read subtemplates from files: %w", err)
	}

	return p.template.Execute(p.writer, p.Data)
}

func (p *Page) RenderText(id, text string) error {
	return p.sse.PatchElements(text, datastar.WithSelectorID(id), datastar.WithModeInner())
}

func TemplatePath(path string) string {
	return templatesHome + "/" + path
}

type Renderer struct {
	Layout    string
	Templates map[string]string
}

func New() *Renderer {
	return &Renderer{
		Layout:    "./templates/index.html",
		Templates: map[string]string{},
	}
}

func (r *Renderer) RegisterTemplate(name, path string) {
	r.Templates[name] = path
}

func (r *Renderer) Render(w http.ResponseWriter, page Page, pageTemplate string, nested ...string) error {
	tmpl, err := template.ParseFiles(r.Layout)
	if err != nil {
		return err
	}

	tmpl, err = tmpl.ParseFiles(r.Templates[pageTemplate])
	if err != nil {
		return err
	}

	if len(nested) > 0 {
		tmpl, err = tmpl.ParseFiles(nested...)
		if err != nil {
			return err
		}
	}

	if err = tmpl.Execute(w, page); err != nil {
		return err
	}

	return nil
}

func init() {
	locVar := os.Getenv("TEMPLATES_HOME")
	if len(locVar) > 0 {
		templatesHome = locVar
	} else {
		templatesHome = "./templates"
	}

	info, err := os.Stat(templatesHome)
	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatalf("'%s' is not a directory", templatesHome)
	}
}
