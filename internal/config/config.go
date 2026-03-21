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

var Config WikiConfig

func init() {
	url, _ := url.Parse("http://localhost:8080")
	Config = WikiConfig{
		Name:        "testwiki",
		Host:        "localhost:8080",
		Port:        8080,
		URL:         url,
		SharedInbox: url.JoinPath("inbox").String(),
	}
}
