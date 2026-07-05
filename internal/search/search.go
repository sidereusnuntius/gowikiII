package search

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

var searchIndexesPath string

type Search struct {
	articles bleve.Index
}

func openOrCreateIndex(name string, mapping mapping.IndexMapping) (bleve.Index, error) {
	var index bleve.Index

	indexPath := filepath.Join(searchIndexesPath, name+".bleve")
	stat, err := os.Stat(indexPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		wikilog.Logger.Debug().Msgf("creating index at %s", indexPath)
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("bleve.New(): failed to create articles index:  %w", err)
		}
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("expected articles index to be at '%s', but there is a file at that path", indexPath)
	} else {
		wikilog.Logger.Debug().Msgf("opening index at %s", indexPath)
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("bleve.Open(): failed to open articles index: %w", err)
		}
	}

	return index, nil
}

func TestSearch() (*Search, error) {
	articleMapping := articleIndexMapping()

	articlesIndex, err := bleve.NewMemOnly(articleMapping)
	if err != nil {
		return nil, err
	}

	return &Search{
		articles: articlesIndex,
	}, nil
}

func Start() (*Search, error) {
	articleMapping := articleIndexMapping()

	articlesIndex, err := openOrCreateIndex("articles", articleMapping)
	if err != nil {
		return nil, err
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
		wikilog.Logger.Fatal().Err(err).Msgf("failed to create directory '%s' to store indexes", searchIndexesPath)
		os.Exit(1)
	}
}
