package activitystreams

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

var space = regexp.MustCompile(`\s+`)

func (o *Object) Slug() (string, bool) {
	slug := string(o.json.GetStringBytes("slug"))
	return slug, len(slug) > 0
}

func (o *Object) Source() (source, mediaType string, ok bool) {
	sourceProperty := o.json.Get("source")
	if sourceProperty == nil {
		return
	}

	source = string(sourceProperty.GetStringBytes("content"))
	mediaType = string(sourceProperty.GetStringBytes("mediaType"))

	ok = len(source) > 0
	return
}

// TODO: in the future we plan to support internationalization, and for that we will handle the source differently.
// The article object itself will have metadata and in its source property will be IRIs through which the
// localized versions of the article can be fetched. We will read the article object, store it in the articles table,
// then fetch each of the mapped localized sources.
func (o *Object) AsArticle() (model.ArticleContent, error) {
	errs := wikierr.NewValidationError()

	id, err := o.Id()
	errs.AppendIfNonNil("id", err)

	var host string
	iri, err := url.Parse(id)
	if err != nil {
		errs.Append("id", errors.New("id is invalid IRI"))
	} else {
		host = iri.Host
	}

	// TODO: lang
	title := string(o.json.GetStringBytes("name"))
	slug, ok := o.Slug()
	if !ok {
		// If slug is not present, we try to create it from title.
		if len(title) == 0 {
			errs.Append("name", wikierr.ErrMissing)
		} else {
			slug = strings.ToLower(
				space.ReplaceAllString(
					strings.TrimSpace(title),
					"",
				),
			)
		}
	}

	// TODO: handle mediaType
	source, _, ok := o.Source()
	if !ok {
		errs.Append("source", wikierr.ErrMissing)
	}

	article := model.ArticleContent{
		Article: model.Article{
			Slug: slug,
			Host: host,
			IRI:  id,
		},
		Title:   title,
		Content: source,
	}

	published, ok, err := o.Published()
	if err != nil {
		errs.AppendIfNonNil("published", err)
	}
	if ok {
		article.Article.Published = published
		article.Published = published
	}

	updated, ok, err := GetTime("updated", o.json)
	if err != nil {
		errs.Append("updated", err)
	}
	if ok {
		article.Updated = updated
	}

	// Optional string fields
	article.Summary = string(o.json.GetStringBytes("summary"))
	article.URL = string(o.json.GetStringBytes("url"))

	return article, errs
}
