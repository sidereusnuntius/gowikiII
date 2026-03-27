package articlesql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sqlhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/sql"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

const (
	selectArticle        = "SELECT id, slug, host, iri, published, federated_edits, restricted_edits FROM articles WHERE slug = ? AND host = ?"
	articleExists        = "SELECT EXISTS(SELECT 1 FROM articles WHERE slug = ? AND host = ?)"
	insertArticle        = "INSERT INTO articles (slug, host, iri, published) VALUES (?, ?, ?, ?) RETURNING id"
	insertArticleContent = `INSERT INTO localized_articles (
	lang_code,
	title,
	content,
	summary,
	url,
	article_id,
	published,
	updated,
	last_fetched
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING id`
	updateArticleContent = `UPDATE localized_articles SET
		content = ?,
		summary = ?,
		updated = ?,
		last_fetched = ?
	WHERE id = ?`
	insertRevision = `INSERT INTO revisions (
		iri,
		code,
		diff,
		prev_revision_id,
		summary,
		localized_article_id,
		published,
		actor_internal_id,
		actor_iri
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`

	// Article query components
	selectArticleContentFulll = `SELECT
		la.id,
		la.lang_code,
		la.title,
		la.summary,
		la.content,
		la.url,
		la.published,
		la.updated,
		la.last_fetched
		FROM localized_articles la JOIN articles a
		ON la.article_id = a.id`

	selectArticleContent = `SELECT
		a.id,
		a.slug,
		a.host,
		a.iri,
		la.id,
		la.title,
		la.lang_code,
		la.summary,
		la.content,
		la.published,
		la.updated,
		la.last_fetched
		FROM articles a JOIN localized_articles la
		ON la.article_id = a.id
		`

	bySlug                 = " WHERE slug = ? AND host = ?"
	byIRI                  = " WHERE iri = ?"
	localized              = " AND lang_code = ?"
	localizedWithFallbacks = "AND lang_code IN (?)"
)

type ArticleStore struct {
	DB *sql.DB
}

func New(db *sql.DB) *ArticleStore {
	return &ArticleStore{
		DB: db,
	}
}

func (as *ArticleStore) UpdateLocalizedArticle(ctx context.Context, content *model.ArticleContent) error {
	_, err := txdb.GetExecutor(ctx, as.DB).ExecContext(ctx, updateArticleContent,
		content.Content,
		sqlhelpers.NullableString(content.Summary),
		sqlhelpers.NullableTimeUnix(content.Updated),
		sqlhelpers.NullableTimeUnix(content.Fetched),
		content.ID,
	)
	return err
}

func (as *ArticleStore) GetArticle(ctx context.Context, slug, host string) (model.Article, error) {
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, selectArticle, slug, host)

	var article model.Article
	var published sql.NullInt64

	err := row.Scan(
		&article.ID,
		&article.Slug,
		&article.Host,
		&article.IRI,
		&published,
		&article.FederatedEdits,
		&article.Restricted,
	)
	if err != nil {
		return model.Article{}, err
	}

	if published.Valid {
		article.Published = time.Unix(published.Int64, 0)
	}

	return article, nil
}

func (as *ArticleStore) ArticleExistsLocally(ctx context.Context, slug, host string) (bool, error) {
	var exists bool
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, articleExists, slug, host)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (as *ArticleStore) SaveArticle(ctx context.Context, article *model.Article) error {
	res, err := txdb.GetExecutor(ctx, as.DB).ExecContext(ctx, insertArticle,
		article.Slug,
		article.Host,
		article.IRI,
		article.Published.Unix(),
	)
	if err != nil {
		return fmt.Errorf("SaveArticle(): %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("SaveArticle(): %w", err)
	}

	article.ID = id
	return nil
}

func (as *ArticleStore) GetArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error) {
	var (
		query strings.Builder
		row   *sql.Row
		err   error
	)

	query.WriteString(selectArticleContent)
	switch {
	case len(req.IRI) > 0:
		query.WriteString(byIRI)
		row = txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, query.String(), req.IRI)
	default:
		query.WriteString(bySlug)
		row = txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, query.String(), req.Slug, req.Host)
	}

	var (
		article   model.ArticleContent
		title     sql.NullString
		summary   sql.NullString
		published sql.NullInt64
		updated   sql.NullInt64
		fetched   sql.NullInt64
	)

	err = row.Scan(
		&article.Article.ID,
		&article.Article.Slug,
		&article.Article.Host,
		&article.Article.IRI,
		&article.ID,
		&title,
		&article.Lang,
		&summary,
		&article.Content,
		&published,
		&updated,
		&fetched,
	)
	if err != nil {
		return article, err
	}

	if title.Valid {
		article.Title = title.String
	}
	if summary.Valid {
		article.Summary = summary.String
	}
	if published.Valid {
		article.Published = time.Unix(published.Int64, 0)
	}
	if updated.Valid {
		article.Updated = time.Unix(updated.Int64, 0)
	}
	if fetched.Valid {
		article.Fetched = time.Unix(fetched.Int64, 0)
	}

	return article, nil
}

func (as *ArticleStore) SaveArticleContent(ctx context.Context, content *model.ArticleContent) error {
	res, err := txdb.GetExecutor(ctx, as.DB).ExecContext(ctx, insertArticleContent,
		content.Lang,
		sqlhelpers.NullableString(content.Title),
		content.Content,
		sqlhelpers.NullableString(content.Summary),
		sqlhelpers.NullableString(content.URL),
		content.Article.ID,
		sqlhelpers.NullableTimeUnix(content.Published),
		sqlhelpers.NullableTimeUnix(content.Updated),
		sqlhelpers.NullableTimeUnix(content.Fetched),
	)
	if err != nil {
		return fmt.Errorf("SaveArticleContent(): %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("res.LastInsertId(): %w", err)
	}

	content.ID = id
	return nil
}

func (as *ArticleStore) SaveRevision(ctx context.Context, revision *model.Revision) error {
	res, err := txdb.GetExecutor(ctx, as.DB).ExecContext(ctx, insertRevision,
		revision.IRI,
		revision.Code,
		revision.Diff,
		sqlhelpers.NullableInt64(revision.Prev),
		sqlhelpers.NullableString(revision.Summary),
		revision.ArticleID,
		revision.Published.Unix(),
		sqlhelpers.NullableInt64(revision.ActorID),
		sqlhelpers.NullableString(revision.ActorIRI),
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	revision.ID = id
	return nil
}
