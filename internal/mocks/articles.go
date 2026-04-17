package mocks

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/stretchr/testify/mock"
)

type MockArticles struct {
	mock.Mock
}

func (ma *MockArticles) RemotePatch(ctx context.Context, patch activitystreams.Patch) error {
	args := ma.Called(ctx, patch)
	return args.Error(0)
}

func (ma *MockArticles) SearchArticles(ctx context.Context, query string) ([]model.Article, error) {
	args := ma.Called(ctx, query)
	return args.Get(0).([]model.Article), args.Error(1)
}

func (ma *MockArticles) ArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error) {
	args := ma.Called(ctx, req)
	return args.Get(0).(model.ArticleContent), args.Error(1)
}

func (ma *MockArticles) LocalEdit(ctx context.Context, in model.ArticleEdit) error {
	args := ma.Called(ctx, in)
	return args.Error(0)
}

type MockArticleStore struct {
	mock.Mock
}

func (mas *MockArticleStore) ArticleExistsLocally(ctx context.Context, iri string) (bool, error) {
	args := mas.Called(ctx, iri)
	return args.Get(0).(bool), args.Error(1)
}

func (mas *MockArticleStore) SaveArticle(ctx context.Context, article *model.Article) error {
	args := mas.Called(ctx, article)
	return args.Error(0)
}

func (mas *MockArticleStore) GetArticle(ctx context.Context, slug, host string) (model.Article, error) {
	args := mas.Called(ctx, slug, host)
	return args.Get(0).(model.Article), args.Error(1)
}

func (mas *MockArticleStore) SearchArticles(ctx context.Context, iris []string) ([]model.Article, error) {
	args := mas.Called(ctx, iris)
	return args.Get(0).([]model.Article), args.Error(1)
}

func (mas *MockArticleStore) GetArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error) {
	args := mas.Called(ctx, req)
	return args.Get(0).(model.ArticleContent), args.Error(1)
}

func (mas *MockArticleStore) SaveArticleContent(ctx context.Context, content *model.ArticleContent) error {
	args := mas.Called(ctx, content)
	return args.Error(0)
}

func (mas *MockArticleStore) UpdateLocalizedArticle(ctx context.Context, content *model.ArticleContent) error {
	args := mas.Called(ctx, content)
	return args.Error(0)
}

func (mas *MockArticleStore) SaveRevision(ctx context.Context, revision *model.Revision) error {
	args := mas.Called(ctx, revision)
	return args.Error(0)
}
