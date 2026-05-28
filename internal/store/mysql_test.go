package store

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"gitlab.meitu.com/xmqibu/internal/content"
)

func TestHasIngestRecordFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"content_slug"}).AddRow("deep-dive")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT content_slug FROM ingest_records WHERE content_type = ? AND external_id = ? LIMIT 1`)).
		WithArgs("article", "ext-1").
		WillReturnRows(rows)

	repo := NewMySQLStore(db)
	found, slug, err := repo.HasIngestRecord(context.Background(), "article", "ext-1")
	if err != nil {
		t.Fatalf("HasIngestRecord() error = %v", err)
	}
	if !found || slug != "deep-dive" {
		t.Fatalf("HasIngestRecord() = (%v, %q), want (true, deep-dive)", found, slug)
	}
}

func TestSaveArticlePersistsArticleTagsAndRecord(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO articles (title, slug, summary, cover_image, body, author, source_links_json, published_at, featured, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("Deep Dive", "deep-dive", "Summary", "", "Body", "Editor", `["https://example.com"]`, now, true, content.StatusPublished).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tags (name, slug) VALUES (?, ?)
ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = VALUES(name)`)).
		WithArgs("ai", "ai").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO content_tags (content_type, content_id, tag_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE tag_id = VALUES(tag_id)`)).
		WithArgs("article", int64(7), int64(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO ingest_records (content_type, external_id, content_slug)
VALUES (?, ?, ?)`)).
		WithArgs("article", "ext-1", "deep-dive").
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	repo := NewMySQLStore(db)
	article, err := repo.SaveArticle(context.Background(), content.IngestArticleInput{
		ExternalID:  "ext-1",
		Title:       "Deep Dive",
		Slug:        "deep-dive",
		Summary:     "Summary",
		Body:        "Body",
		Author:      "Editor",
		SourceLinks: []string{"https://example.com"},
		Tags:        []string{"ai"},
		PublishedAt: now,
		Featured:    true,
	})
	if err != nil {
		t.Fatalf("SaveArticle() error = %v", err)
	}
	if article.ID != 7 || article.Slug != "deep-dive" {
		t.Fatalf("SaveArticle() = %#v, want saved article", article)
	}
}

func TestListArticlesReturnsPublishedArticles(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "title", "slug", "summary", "cover_image", "body", "author", "source_links_json", "published_at", "featured", "status",
	}).AddRow(1, "Main Story", "main-story", "Summary", "", "Body", "Editor", `["https://example.com"]`, now, true, content.StatusPublished)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, title, slug, summary, cover_image, body, author, source_links_json, published_at, featured, status
FROM articles
WHERE status = ?
ORDER BY featured DESC, published_at DESC
LIMIT ?`)).
		WithArgs(content.StatusPublished, 6).
		WillReturnRows(rows)

	repo := NewMySQLStore(db)
	articles, err := repo.ListArticles(context.Background(), 6)
	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}
	if len(articles) != 1 || articles[0].Slug != "main-story" {
		t.Fatalf("ListArticles() = %#v, want article", articles)
	}
}
