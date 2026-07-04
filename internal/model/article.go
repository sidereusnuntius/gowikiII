package model

import "time"

type ArticleFilter struct {
	Query    string
	NextPage int64
}

type SearchResults struct {
	Query   string
	Results []SearchResult

	HasNextPage bool
	Next        int64
	Total       int
}

type SearchResult struct {
	URL     string
	Summary string
	Title   string
}

type Article struct {
	ID             int64
	Slug           string
	Host           string
	IRI            string
	Published      time.Time
	FederatedEdits bool
	// Restricted indicates whether the article only accepts
	// edits from trusted users.
	Restricted bool
}

type ArticleContent struct {
	Article   Article   `json:"article"`
	ID        int64     `json:"id"`
	Lang      string    `json:"lang"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Summary   string    `json:"summary"`
	URL       string    `json:"url"`
	Published time.Time `json:"published"`
	Updated   time.Time `json:"updated"`
	Fetched   time.Time `json:"fetched"`
}

type Revision struct {
	// Code is a unique code that identifies the revision.
	Code      string
	ID        int64
	IRI       string
	Diff      string
	Summary   string
	Prev      int64
	ArticleID int64
	Published time.Time
	ActorID   int64
	ActorIRI  string
}

type ArticleRequest struct {
	Slug  string
	Host  string
	IRI   string
	Langs []string
	IDs   []string
}

type ArticleEdit struct {
	ActorID    int64
	ActorIRI   string
	LocalActor bool
	Slug       string
	Host       string
	IRI        string
	Lang       string
	NewContent string
	Summary    string
	// LastEdit is the timestamp of the last revision of an article at the moment in which the user started editing it.
	// This is used when an edit happens between the user starting to edit an article and trying to save their changes.
	LastEdit time.Time
}
