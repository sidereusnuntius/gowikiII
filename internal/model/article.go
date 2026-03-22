package model

import "time"

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
	Article   Article
	ID        int64
	Lang      string
	Title     string
	Content   string
	Summary   string
	URL       string
	Published time.Time
	Updated   time.Time
	Fetched   time.Time
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
	Slug       string
	Host       string
	IRI        string
	Lang       string
	NewContent string
	// LastEdit is the timestamp of the last revision of an article at the moment in which the user started editing it.
	// This is used when an edit happens between the user starting to edit an article and trying to save their changes.
	LastEdit time.Time
}
