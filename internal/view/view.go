package view

import "time"

type ArticleHeader struct {
	ArticleURL string
	EditURL string
	HistoryURL string
}

type Editor struct {
	ArticleHeader
	Slug string
	Host string
	Content string
	EditSummary string
	LastModified time.Time
	ActionURL string
}

type Article struct {
	ArticleHeader
	Slug string
	Host string
	Content string
}
