package streams

import (
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Article struct {
	Base
	Slug         string `json:"slug,omitempty"`
	Name         string `json:"name,omitempty"`
	Edits        string `json:"edits,omitempty"`
	AttributedTo string `json:"attributedTo,omitempty"`
	Protected    bool   `json:"protected"`
	Content      string `json:"content,omitempty"`
	Source       Source `json:"source,omitzero"`
}

type Source struct {
	Content   string `json:"content,omitempty"`
	Mediatype string `json:"mediaType,omitempty"`
}

func ArticleAS(article *model.ArticleContent) Article {
	var published, updated string
	if !article.Published.IsZero() {
		published = article.Published.Format(Format)
	}
	if !article.Updated.IsZero() {
		updated = article.Updated.Format(Format)
	}
	return Article{
		Base: Base{
			Context:   context,
			Type:      "Article",
			Id:        article.Article.IRI,
			Published: published,
			Updated:   updated,
			Url:       article.URL,
		},
		Slug: article.Article.Slug,
		Name: article.Title,
		// Edits: article., TODO
		AttributedTo: fmt.Sprintf("http://%s", article.Article.Host),
		// Protected: , TODO
		Content: article.Content, // TODO: render Markdown
		Source: Source{
			Content: article.Content,
			// Mediatype: , TODO
		},
	}
}
