package view

import (
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/config"
)

const revisionTimeFormat = "2006-01-02 15:04"

func RevisionURL(id int64) string {
	return fmt.Sprintf("/revision/%d", id)
}

func Username(config *config.WikiConfig, username, host string) string {
	if host == "" || host == config.Host {
		return username
	}

	return fmt.Sprintf("%s@%s", username, host)
}

func ActorURL(config *config.WikiConfig, username, host string) string {
	if host == "" || host == config.Host {
		return "/u/" + username
	}

	return fmt.Sprintf("/u/%s@%s", username, host)
}

func ArticleURL(config *config.WikiConfig, slug, host string) string {
	if host == "" || host == config.Host {
		return "/a/" + slug
	}

	return fmt.Sprintf("/a/%s/%s", host, slug)
}

func ArticleEditURL(config *config.WikiConfig, slug, host string) string {
	if host == "" || host == config.Host {
		return fmt.Sprintf("/a/%s/edit", slug)
	}

	return fmt.Sprintf("/a/%s/%s/edit", host, slug)
}

func ArticleTitle(config *config.WikiConfig, slug, host string) string {
	if host == "" || host == config.Host {
		return slug
	}

	return fmt.Sprintf("%s@%s", slug, host)
}
