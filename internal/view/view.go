package view

import (
	"html/template"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Editor struct {
	Slug         string
	Host         string
	Content      string
	Preview      string
	EditSummary  string
	LastModified time.Time
	ActionURL    string
}

type Article struct {
	Slug    string
	Host    string
	Content string
}

type RevisionView struct {
	Published        string
	PublishedDisplay string
	Actor            string
	ActorURL         string
	Summary          string
	Diff             string
	DiffHTML         template.HTML
	URL              string
	Article          string
	ArticleURL       string
}

func HistoryView(config *config.WikiConfig, revisions []model.Revision, general bool) History {
	history := History{
		Revisions: make([]RevisionView, len(revisions)),
	}

	for i := range revisions {
		history.Revisions[i] = RevisionView{
			Published:        revisions[i].Published.Format("2006-01-02"),
			PublishedDisplay: revisions[i].Published.Format(revisionTimeFormat),
			Actor:            Username(config, revisions[i].ActorUsername, revisions[i].ActorHost),
			ActorURL:         ActorURL(config, revisions[i].ActorUsername, revisions[i].ActorHost),
			Summary:          revisions[i].Summary,
			URL:              RevisionURL(revisions[i].ID),
		}

		if general {
			history.Revisions[i].Article = ArticleTitle(config, revisions[i].ArticleSlug, revisions[i].ArticleHost)
			history.Revisions[i].ArticleURL = ArticleURL(config, revisions[i].ArticleSlug, revisions[i].ArticleHost)
		}
	}

	return history
}

type History struct {
	Revisions []RevisionView
}
