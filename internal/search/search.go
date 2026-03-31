package search

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

var searchIndexesPath string

type Search struct {
	articles bleve.Index
}

func Start() (*Search, error) {
	articleMapping := articleIndexMapping()

	var articlesIndex bleve.Index
	articlesIndexPath := filepath.Join(searchIndexesPath, "articles.bleve")
	stat, err := os.Stat(articlesIndexPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		wikilog.Logger.Debug().Msgf("creating index at %s", articlesIndexPath)
		articlesIndex, err = bleve.New(articlesIndexPath, articleMapping)
		if err != nil {
			return nil, fmt.Errorf("bleve.New(): failed to create articles index:  %w", err)
		}
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("expected articles index to be at '%s', but there is a file at that path")
	} else {
		wikilog.Logger.Debug().Msgf("opening index at %s", articlesIndexPath)
		articlesIndex, err = bleve.Open(articlesIndexPath)
		if err != nil {
			return nil, fmt.Errorf("bleve.Open(): failed to open articles index: %w", err)
		}
	}

	return &Search{
		articles: articlesIndex,
	}, nil
}

func (s *Search) Close() {
	err := s.articles.Close()
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to close articles index")
	}
}
func (s *Search) SearchArticles(query string) (*bleve.SearchResult, error) {
	q := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequest(q)

	return s.articles.Search(req)
}

func (s *Search) IndexArticle(article *model.ArticleContent) error {
	wikilog.Logger.Info().Msgf("indexing article %s", article.Article.IRI)
	return s.articles.Index(article.Article.IRI, article)
}

func init() {
	if envSearchIndexesDir := os.Getenv("SEARCH_INDEXES_DIR"); envSearchIndexesDir != "" {
		searchIndexesPath = envSearchIndexesDir
	} else {
		searchIndexesPath = "./indexes"
	}

	err := os.MkdirAll(searchIndexesPath, os.ModePerm)
	if err != nil {
		wikilog.Logger.Fatal().Err(err).Msgf("failed to create directory '%s' to store indexes")
		os.Exit(1)
	}
}
