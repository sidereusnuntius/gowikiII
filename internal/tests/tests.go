package tests

import (
	"context"
	"database/sql"
	"net/url"
	"strconv"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

func TestConfig(addr string) config.WikiConfig {
	url, _ := url.Parse(addr)
	port, _ := strconv.Atoi(url.Port())
	return config.WikiConfig{
		Name:        "testwiki",
		Host:        url.Host,
		Port:        port,
		URL:         url,
		SharedInbox: url.JoinPath("inbox").String(),
	}
}

func TestDB(ctx context.Context) (*sql.DB, error) {
	config := config.DbConfig{
		URL: ":memory:",
	}

	return db.Open(ctx, config)
}
