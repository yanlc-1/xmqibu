package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gitlab.meitu.com/xmqibu/internal/content"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) HasIngestRecord(ctx context.Context, kind, externalID string) (bool, string, error) {
	var slug string
	err := s.db.QueryRowContext(ctx, `SELECT content_slug FROM ingest_records WHERE content_type = ? AND external_id = ? LIMIT 1`, kind, externalID).Scan(&slug)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, slug, nil
}

func (s *MySQLStore) SaveArticle(ctx context.Context, input content.IngestArticleInput) (content.Article, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return content.Article{}, err
	}
	defer tx.Rollback()

	sourceLinksJSON, err := json.Marshal(input.SourceLinks)
	if err != nil {
		return content.Article{}, err
	}

	result, err := tx.ExecContext(ctx, `INSERT INTO articles (title, slug, summary, cover_image, body, author, source_links_json, published_at, featured, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Title,
		input.Slug,
		input.Summary,
		input.CoverImage,
		input.Body,
		input.Author,
		string(sourceLinksJSON),
		input.PublishedAt,
		input.Featured,
		content.StatusPublished,
	)
	if err != nil {
		return content.Article{}, err
	}
	articleID, err := result.LastInsertId()
	if err != nil {
		return content.Article{}, err
	}

	if err := s.attachTags(ctx, tx, "article", articleID, input.Tags); err != nil {
		return content.Article{}, err
	}
	if err := s.recordIngest(ctx, tx, "article", input.ExternalID, input.Slug); err != nil {
		return content.Article{}, err
	}
	if err := tx.Commit(); err != nil {
		return content.Article{}, err
	}

	return content.Article{
		ID:          articleID,
		Title:       input.Title,
		Slug:        input.Slug,
		Summary:     input.Summary,
		CoverImage:  input.CoverImage,
		Body:        input.Body,
		Author:      input.Author,
		SourceLinks: append([]string(nil), input.SourceLinks...),
		PublishedAt: input.PublishedAt,
		Featured:    input.Featured,
		Status:      content.StatusPublished,
	}, nil
}

func (s *MySQLStore) SaveBrief(ctx context.Context, input content.IngestBriefInput) (content.Brief, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return content.Brief{}, err
	}
	defer tx.Rollback()

	sourceLinksJSON, err := json.Marshal(input.SourceLinks)
	if err != nil {
		return content.Brief{}, err
	}

	result, err := tx.ExecContext(ctx, `INSERT INTO briefs (title, slug, summary, external_url, importance, source_links_json, published_at, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Title,
		input.Slug,
		input.Summary,
		input.ExternalURL,
		input.Importance,
		string(sourceLinksJSON),
		input.PublishedAt,
		content.StatusPublished,
	)
	if err != nil {
		return content.Brief{}, err
	}
	briefID, err := result.LastInsertId()
	if err != nil {
		return content.Brief{}, err
	}

	if err := s.attachTags(ctx, tx, "brief", briefID, input.Tags); err != nil {
		return content.Brief{}, err
	}
	if err := s.recordIngest(ctx, tx, "brief", input.ExternalID, input.Slug); err != nil {
		return content.Brief{}, err
	}
	if err := tx.Commit(); err != nil {
		return content.Brief{}, err
	}

	return content.Brief{
		ID:          briefID,
		Title:       input.Title,
		Slug:        input.Slug,
		Summary:     input.Summary,
		ExternalURL: input.ExternalURL,
		Importance:  input.Importance,
		SourceLinks: append([]string(nil), input.SourceLinks...),
		PublishedAt: input.PublishedAt,
		Status:      content.StatusPublished,
	}, nil
}

func (s *MySQLStore) ListArticles(ctx context.Context, limit int) ([]content.Article, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, slug, summary, cover_image, body, author, source_links_json, published_at, featured, status
FROM articles
WHERE status = ?
ORDER BY featured DESC, published_at DESC
LIMIT ?`, content.StatusPublished, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanArticles(rows)
}

func (s *MySQLStore) ListBriefs(ctx context.Context, page, pageSize int) ([]content.Brief, bool, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, slug, summary, external_url, importance, source_links_json, published_at, status
FROM briefs
WHERE status = ?
ORDER BY published_at DESC
LIMIT ? OFFSET ?`, content.StatusPublished, pageSize+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	items, err := scanBriefs(rows)
	if err != nil {
		return nil, false, err
	}
	return trimBriefsPage(items, pageSize), len(items) > pageSize, nil
}

func (s *MySQLStore) ListTopTags(ctx context.Context, limit int) ([]content.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.name, t.slug, COUNT(ct.tag_id) AS usage_count
FROM tags t
JOIN content_tags ct ON ct.tag_id = t.id
GROUP BY t.id, t.name, t.slug
ORDER BY usage_count DESC, t.name ASC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTags(rows)
}

func (s *MySQLStore) GetArticleBySlug(ctx context.Context, slug string) (content.Article, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, title, slug, summary, cover_image, body, author, source_links_json, published_at, featured, status
FROM articles
WHERE slug = ? AND status = ?
LIMIT 1`, slug, content.StatusPublished)
	article, err := scanArticleRow(row)
	if err != nil {
		return content.Article{}, err
	}
	tags, err := s.loadTagsForContent(ctx, "article", article.ID)
	if err != nil {
		return content.Article{}, err
	}
	article.Tags = tags
	return article, nil
}

func (s *MySQLStore) GetTagBySlug(ctx context.Context, slug string) (content.Tag, error) {
	var tag content.Tag
	err := s.db.QueryRowContext(ctx, `SELECT id, name, slug FROM tags WHERE slug = ? LIMIT 1`, slug).Scan(&tag.ID, &tag.Name, &tag.Slug)
	return tag, err
}

func (s *MySQLStore) ListArticlesByTag(ctx context.Context, slug string, limit int) ([]content.Article, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id, a.title, a.slug, a.summary, a.cover_image, a.body, a.author, a.source_links_json, a.published_at, a.featured, a.status
FROM articles a
JOIN content_tags ct ON ct.content_type = 'article' AND ct.content_id = a.id
JOIN tags t ON t.id = ct.tag_id
WHERE t.slug = ? AND a.status = ?
ORDER BY a.featured DESC, a.published_at DESC
LIMIT ?`, slug, content.StatusPublished, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *MySQLStore) ListBriefsByTag(ctx context.Context, slug string, page, pageSize int) ([]content.Brief, bool, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx, `SELECT b.id, b.title, b.slug, b.summary, b.external_url, b.importance, b.source_links_json, b.published_at, b.status
FROM briefs b
JOIN content_tags ct ON ct.content_type = 'brief' AND ct.content_id = b.id
JOIN tags t ON t.id = ct.tag_id
WHERE t.slug = ? AND b.status = ?
ORDER BY b.published_at DESC
LIMIT ? OFFSET ?`, slug, content.StatusPublished, pageSize+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	items, err := scanBriefs(rows)
	if err != nil {
		return nil, false, err
	}
	return trimBriefsPage(items, pageSize), len(items) > pageSize, nil
}

func (s *MySQLStore) Search(ctx context.Context, term string, page, pageSize int) ([]content.Article, []content.Brief, []content.Tag, bool, error) {
	pattern := "%" + strings.TrimSpace(term) + "%"
	articleRows, err := s.db.QueryContext(ctx, `SELECT DISTINCT a.id, a.title, a.slug, a.summary, a.cover_image, a.body, a.author, a.source_links_json, a.published_at, a.featured, a.status
FROM articles a
LEFT JOIN content_tags ct ON ct.content_type = 'article' AND ct.content_id = a.id
LEFT JOIN tags t ON t.id = ct.tag_id
WHERE a.status = ? AND (a.title LIKE ? OR a.summary LIKE ? OR t.name LIKE ?)
ORDER BY a.featured DESC, a.published_at DESC
LIMIT ? OFFSET ?`, content.StatusPublished, pattern, pattern, pattern, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, nil, nil, false, err
	}
	defer articleRows.Close()

	articles, err := scanArticles(articleRows)
	if err != nil {
		return nil, nil, nil, false, err
	}

	briefRows, err := s.db.QueryContext(ctx, `SELECT DISTINCT b.id, b.title, b.slug, b.summary, b.external_url, b.importance, b.source_links_json, b.published_at, b.status
FROM briefs b
LEFT JOIN content_tags ct ON ct.content_type = 'brief' AND ct.content_id = b.id
LEFT JOIN tags t ON t.id = ct.tag_id
WHERE b.status = ? AND (b.title LIKE ? OR b.summary LIKE ? OR t.name LIKE ?)
ORDER BY b.published_at DESC
LIMIT ? OFFSET ?`, content.StatusPublished, pattern, pattern, pattern, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, nil, nil, false, err
	}
	defer briefRows.Close()

	briefs, err := scanBriefs(briefRows)
	if err != nil {
		return nil, nil, nil, false, err
	}

	tagRows, err := s.db.QueryContext(ctx, `SELECT id, name, slug, 0 AS usage_count
FROM tags
WHERE name LIKE ? OR slug LIKE ?
ORDER BY name ASC
LIMIT 8`, pattern, pattern)
	if err != nil {
		return nil, nil, nil, false, err
	}
	defer tagRows.Close()

	tags, err := scanTags(tagRows)
	if err != nil {
		return nil, nil, nil, false, err
	}

	hasMore := len(articles) > pageSize || len(briefs) > pageSize
	return trimArticlesPage(articles, pageSize), trimBriefsPage(briefs, pageSize), tags, hasMore, nil
}

func (s *MySQLStore) attachTags(ctx context.Context, tx *sql.Tx, contentType string, contentID int64, tags []string) error {
	for _, tag := range tags {
		result, err := tx.ExecContext(ctx, `INSERT INTO tags (name, slug) VALUES (?, ?)
ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = VALUES(name)`, tag, content.Slugify(tag))
		if err != nil {
			return err
		}
		tagID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO content_tags (content_type, content_id, tag_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE tag_id = VALUES(tag_id)`, contentType, contentID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) recordIngest(ctx context.Context, tx *sql.Tx, contentType, externalID, slug string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO ingest_records (content_type, external_id, content_slug)
VALUES (?, ?, ?)`, contentType, externalID, slug)
	return err
}

func (s *MySQLStore) loadTagsForContent(ctx context.Context, contentType string, contentID int64) ([]content.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.name, t.slug, 0 AS usage_count
FROM tags t
JOIN content_tags ct ON ct.tag_id = t.id
WHERE ct.content_type = ? AND ct.content_id = ?
ORDER BY t.name ASC`, contentType, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTags(rows)
}

func scanArticles(rows *sql.Rows) ([]content.Article, error) {
	var items []content.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, article)
	}
	return items, rows.Err()
}

func scanArticleRow(row *sql.Row) (content.Article, error) {
	var (
		article         content.Article
		sourceLinksJSON string
	)
	err := row.Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Summary,
		&article.CoverImage,
		&article.Body,
		&article.Author,
		&sourceLinksJSON,
		&article.PublishedAt,
		&article.Featured,
		&article.Status,
	)
	if err != nil {
		return content.Article{}, err
	}
	article.SourceLinks = decodeSourceLinks(sourceLinksJSON)
	return article, nil
}

func scanArticle(rows *sql.Rows) (content.Article, error) {
	var (
		article         content.Article
		sourceLinksJSON string
	)
	err := rows.Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Summary,
		&article.CoverImage,
		&article.Body,
		&article.Author,
		&sourceLinksJSON,
		&article.PublishedAt,
		&article.Featured,
		&article.Status,
	)
	if err != nil {
		return content.Article{}, err
	}
	article.SourceLinks = decodeSourceLinks(sourceLinksJSON)
	return article, nil
}

func scanBriefs(rows *sql.Rows) ([]content.Brief, error) {
	var items []content.Brief
	for rows.Next() {
		var (
			item            content.Brief
			sourceLinksJSON string
		)
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Slug,
			&item.Summary,
			&item.ExternalURL,
			&item.Importance,
			&sourceLinksJSON,
			&item.PublishedAt,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}
		item.SourceLinks = decodeSourceLinks(sourceLinksJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanTags(rows *sql.Rows) ([]content.Tag, error) {
	var items []content.Tag
	for rows.Next() {
		var item content.Tag
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.UsageCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func decodeSourceLinks(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var links []string
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		return []string{fmt.Sprintf("invalid:%s", raw)}
	}
	return links
}

func trimArticlesPage(items []content.Article, pageSize int) []content.Article {
	if len(items) > pageSize {
		return items[:pageSize]
	}
	return items
}

func trimBriefsPage(items []content.Brief, pageSize int) []content.Brief {
	if len(items) > pageSize {
		return items[:pageSize]
	}
	return items
}
