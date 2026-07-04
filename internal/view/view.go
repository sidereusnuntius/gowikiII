package view

import "time"

type PageControl struct {
	Selected bool
	URL      string
	Label    string
}

type Editor struct {
	Controls     [3]PageControl
	Slug         string
	Host         string
	Content      string
	EditSummary  string
	LastModified time.Time
	ActionURL    string
}

type Article struct {
	Controls [3]PageControl
	Slug     string
	Host     string
	Content  string
}
