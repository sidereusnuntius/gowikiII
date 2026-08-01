package articles

import (
	"context"
	"fmt"
	"strings"

	wikilink "github.com/sidereusnuntius/goldmark-wikilink"
	"github.com/sidereusnuntius/gowiki/internal/view"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

func (as *ArticleService) ResolveWikilink(n *wikilink.Node) (destination []byte, classes []string, err error) {
	target := string(n.Target)

	if !strings.Contains(target, ":") {
		class := "red-link"

		var exists bool
		exists, err = as.Store.ArticleSlugExists(context.Background(), target)
		if err != nil {
			wikilog.Logger.Error().Err(err).Str("target", target).Msg("ArticleService.ResolveWikiLink: failed to check if slug existence")
		} else if exists {
			class = "blue-link"
		}

		classes = append(classes, class)
		destination = []byte(view.ArticleURL(&as.Config, target, ""))
		return
	}

	substrs := strings.SplitN(target, ":", 1)
	if len(substrs) != 2 {
		err = fmt.Errorf("invalid target: %s", target)
		return
	}

	class := "red-link"
	exists, err := as.Store.ExternalArticleExistsLocally(context.Background(), substrs[1], substrs[0])
	if err != nil {
		wikilog.Logger.Error().Err(err).Str("target", target).Msg("ArticleService.ResolveWikiLink: failed to check if slug existence")
	} else if exists {
		class = "blue-link"
	}

	classes = append(classes, class)
	destination = []byte(view.ArticleURL(&as.Config, substrs[1], substrs[0]))

	return
}
