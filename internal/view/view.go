package view

import (
	"time"
)

type Editor struct {
	Slug         string
	Host         string
	Content      string
	EditSummary  string
	LastModified time.Time
	ActionURL    string
}

type Article struct {
	Slug    string
	Host    string
	Content string
}
