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
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

const (
	selectSlugExists = "SELECT EXISTS(SELECT 1 FROM articles WHERE slug = ? LIMIT 1)"
	selectArticle    = "SELECT id, slug, host, iri, published, federated_edits, restricted_edits FROM articles WHERE slug = ? AND host = ?"
	selectArticles   = "SELECT id, slug, host, iri FROM articles where iri IN (%s)"
	// selectLocalizedArticleID = "SELECT la.id FROM localized_articles la JOIN articles a ON la.article_id = a.id WHERE a.slug = ? AND a.host = ? AND la.lang_code = ?"
	selectLocalizedArticleID = "SELECT la.id FROM localized_articles la JOIN articles a ON la.article_id = a.id WHERE a.slug = ? AND a.host = ?"
	articleExists            = "SELECT EXISTS(SELECT 1 FROM articles WHERE iri = ?)"
	externalArticleExists    = "SELECT EXISTS(SELECT 1 FROM articles WHERE host = ? AND slug = ?)"
	insertArticle            = "INSERT INTO articles (slug, host, iri, published) VALUES (?, ?, ?, ?) RETURNING id"
	insertArticleContent     = `INSERT INTO localized_articles (
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
		reverse_diff,
		prev_revision_id,
		summary,
		localized_article_id,
		published,
		actor_internal_id,
		actor_iri
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`

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
	selectRevisionsTpl = `SELECT
			r.id,
			r.code,
			r.summary,
			r.published,
			art.slug,
			art.host,
			la.url,
			a.id,
			a.username,
			a.host
		FROM
			revisions r
		JOIN actors a
		ON
			r.actor_internal_id = a.id
		JOIN localized_articles la
		ON r.localized_article_id = la.id
		JOIN articles art
		ON la.article_id = art.id`
	selectRevisions = selectRevisionsTpl + `
		WHERE
			r.localized_article_id = ? AND r.published <= ?
		ORDER BY r.published DESC LIMIT ?`
	selectRecentRevisions = selectRevisionsTpl + `
		WHERE
			r.published <= ?
		ORDER BY r.published DESC LIMIT ?`
	selectRevision = `SELECT
			r.id,
			r.code,
			r.summary,
			r.published,
			r.diff,
			art.slug,
			art.host,
			la.id,
			art.iri,
			a.id,
			a.username,
			a.host
		FROM
			revisions r
		JOIN actors a
		ON
			r.actor_internal_id = a.id
		JOIN localized_articles la
		ON r.localized_article_id = la.id
		JOIN articles art
		ON la.article_id = art.id
		WHERE
			r.id = ?`
	selectRevisionsReverseDiffs = `select reverse_diff from revisions where localized_article_id = ? and id >= ? order by id desc`

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

func (as *ArticleStore) SearchArticles(ctx context.Context, iris []string, filter model.ArticleFilter) ([]model.Article, error) {
	var builder strings.Builder
	for i, iri := range iris {
		builder.WriteRune('\'')
		builder.WriteString(iri)
		builder.WriteRune('\'')
		if i != len(iris)-1 {
			builder.WriteRune(',')
		}
	}
	fmt.Println(builder.String())

	query := selectArticles
	if filter.NextPage > 0 {
		query += fmt.Sprintf(" AND id >= %d", filter.NextPage)
	}
	query += " LIMIT 20"

	rows, err := txdb.GetExecutor(ctx, as.DB).QueryContext(ctx,
		fmt.Sprintf(query, builder.String()),
	)
	if err != nil {
		return nil, err
	}

	result := make([]model.Article, 0)
	var article model.Article
	for rows.Next() {
		err := rows.Scan(
			&article.ID,
			&article.Slug,
			&article.Host,
			&article.IRI,
		)
		fmt.Println("article:", article)
		if err != nil {
			return nil, err
		}

		result = append(result, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	fmt.Println(result)

	return result, nil
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

func (as *ArticleStore) GetLocalizedArticleID(ctx context.Context, slug, host, lang string) (int64, error) {
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, selectLocalizedArticleID, slug, host)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, sqlhelpers.HandleErr(err)
	}

	return id, nil
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
		return model.Article{}, sqlhelpers.HandleErr(err)
	}

	if published.Valid {
		article.Published = time.Unix(published.Int64, 0)
	}

	return article, nil
}

func (as *ArticleStore) ArticleSlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, selectSlugExists, slug)
	if err := row.Scan(&exists); err != nil {
		return false, sqlhelpers.HandleErr(err)
	}

	return exists, nil
}

func (as *ArticleStore) ExternalArticleExistsLocally(ctx context.Context, slug, host string) (bool, error) {
	var exists bool
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, externalArticleExists, host, slug)
	if err := row.Scan(&exists); err != nil {
		return false, sqlhelpers.HandleErr(err)
	}

	return exists, nil
}

func (as *ArticleStore) ArticleExistsLocally(ctx context.Context, iri string) (bool, error) {
	var exists bool
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, articleExists, iri)
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
		fmt.Println("IRI:", req.IRI)
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
		wikilog.Logger.Error().Err(err).Any("req", req).Send()
		return article, sqlhelpers.HandleErr(err)
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
		revision.ReverseDiff,
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

func (as *ArticleStore) RevisionByID(ctx context.Context, revisionID int64) (model.Revision, error) {
	fmt.Printf("%d", revisionID)
	row := txdb.GetExecutor(ctx, as.DB).QueryRowContext(ctx, selectRevision, revisionID)
	var (
		revision  model.Revision
		summary   sql.NullString
		published int64
	)

	err := row.Scan(
		&revision.ID,
		&revision.Code,
		&summary,
		&published,
		&revision.Diff,
		&revision.ArticleSlug,
		&revision.ArticleHost,
		&revision.ArticleID,
		&revision.ArticleURL,
		&revision.ActorID,
		&revision.ActorUsername,
		&revision.ActorHost,
	)
	if err != nil {
		return revision, sqlhelpers.HandleErr(err)
	}

	if summary.Valid {
		revision.Summary = summary.String
	}
	revision.Published = time.Unix(published, 0)

	return revision, nil
}

func (as *ArticleStore) RecentChanges(ctx context.Context, after time.Time, limit int) ([]model.Revision, error) {
	if after.IsZero() {
		after = time.Now()
	}

	rows, err := txdb.GetExecutor(ctx, as.DB).QueryContext(ctx, selectRecentRevisions, after.Unix(), limit)
	if err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	var revisions []model.Revision
	for rows.Next() {
		r, err := as.scanRevision(rows)
		if err != nil {
			return nil, sqlhelpers.HandleErr(err)
		}

		revisions = append(revisions, r)
	}
	if err = rows.Err(); err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	return revisions, nil
}

func (as *ArticleStore) RevisionHistory(ctx context.Context, localizedArticleID int64, after time.Time, limit int) ([]model.Revision, error) {
	if after.IsZero() {
		after = time.Now().UTC()
	}
	rows, err := txdb.GetExecutor(ctx, as.DB).QueryContext(ctx, selectRevisions, localizedArticleID, after.Unix(), limit)
	if err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	var (
		revisions    []model.Revision
		r            model.Revision
		published    int64
		summary, url sql.NullString
	)
	for rows.Next() {
		err = rows.Scan(
			&r.ID,
			&r.Code,
			&summary,
			&published,
			&r.ArticleSlug,
			&r.ArticleHost,
			&url,
			&r.ActorID,
			&r.ActorUsername,
			&r.ActorHost,
		)
		if err != nil {
			return nil, sqlhelpers.HandleErr(err)
		}

		if summary.Valid {
			r.Summary = summary.String
		}
		if url.Valid {
			r.ArticleURL = url.String
		}
		r.Published = time.Unix(published, 0)
		revisions = append(revisions, r)
	}
	if err = rows.Err(); err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	return revisions, nil
}

func (as *ArticleStore) ArticleReverseHistory(ctx context.Context, localizedArticleID, targetRevisionID int64) ([]string, error) {
	wikilog.Logger.Info().Msgf("fetching history: %d %d", localizedArticleID, targetRevisionID)
	rows, err := txdb.GetExecutor(ctx, as.DB).QueryContext(ctx, selectRevisionsReverseDiffs, localizedArticleID, targetRevisionID)
	if err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	var diff string
	diffs := make([]string, 0)
	for rows.Next() {
		if err := rows.Scan(&diff); err != nil {
			return nil, sqlhelpers.HandleErr(err)
		}

		diffs = append(diffs, diff)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlhelpers.HandleErr(err)
	}

	return diffs, nil
}

func (as *ArticleStore) scanRevision(row sqlhelpers.Scanner) (model.Revision, error) {
	var (
		r            model.Revision
		published    int64
		summary, url sql.NullString
	)
	err := row.Scan(
		&r.ID,
		&r.Code,
		&summary,
		&published,
		&r.ArticleSlug,
		&r.ArticleHost,
		&url,
		&r.ActorID,
		&r.ActorUsername,
		&r.ActorHost,
	)
	if err != nil {
		return r, err
	}

	if summary.Valid {
		r.Summary = summary.String
	}
	if url.Valid {
		r.ArticleURL = url.String
	}
	r.Published = time.Unix(published, 0)

	return r, nil
}
