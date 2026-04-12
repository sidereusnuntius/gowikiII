package config

import (
	"net/url"
)

type WikiConfig struct {
	Name        string
	Host        string
	Port        int
	URL         *url.URL
	SharedInbox string
}

type DbConfig struct {
	URL     string
	Timeout int
	// TODO: add other options
}
