package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.meitu.com/xmqibu/internal/content"
)

type ingestStub struct {
	articleResult IngestResponse
	briefResult   IngestResponse
	articleErr    error
	briefErr      error
}

func (s ingestStub) IngestArticle(ctx context.Context, input content.IngestArticleInput) (IngestResponse, error) {
	return s.articleResult, s.articleErr
}

func (s ingestStub) IngestBrief(ctx context.Context, input content.IngestBriefInput) (IngestResponse, error) {
	return s.briefResult, s.briefErr
}

type queryStub struct {
	home    content.HomePageData
	article content.Article
	briefs  BriefPageData
	tagPage TagPageData
	search  SearchPageData
	err     error
}

func (s queryStub) HomePage(ctx context.Context) (content.HomePageData, error) {
	return s.home, s.err
}

func (s queryStub) ArticleBySlug(ctx context.Context, slug string) (content.Article, error) {
	if slug == "" {
		return content.Article{}, errors.New("missing slug")
	}
	return s.article, s.err
}

func (s queryStub) Briefs(ctx context.Context, page, pageSize int) (BriefPageData, error) {
	return s.briefs, s.err
}

func (s queryStub) TagPage(ctx context.Context, slug string, page, pageSize int) (TagPageData, error) {
	return s.tagPage, s.err
}

func (s queryStub) Search(ctx context.Context, term string, page, pageSize int) (SearchPageData, error) {
	return s.search, s.err
}

func newTestApp(t *testing.T, ingest IngestService, query QueryService) *App {
	t.Helper()

	tmpl := template.Must(template.New("base").Parse(`
		{{define "home"}}<h1>{{.HeroArticle.Title}}</h1>{{if .HeroArticle.CoverImage}}<img src="{{.HeroArticle.CoverImage}}" alt="hero cover">{{end}}{{range .Sections}}{{range .Briefs}}{{if .ExternalURL}}<a href="{{.ExternalURL}}">{{.Title}}</a>{{end}}{{end}}<section>{{.Label}}</section>{{end}}{{end}}
		{{define "article"}}<article><h1>{{.Article.Title}}</h1>{{if .Article.CoverImage}}<img src="{{.Article.CoverImage}}" alt="story cover">{{end}}<div>{{.Article.Body}}</div></article>{{end}}
		{{define "briefs"}}<section>{{range .Briefs}}<article>{{if .ExternalURL}}<a href="{{.ExternalURL}}">{{.Title}}</a>{{else}}{{.Title}}{{end}}</article>{{end}}</section>{{end}}
		{{define "tag"}}<section><h1>{{.Tag.Name}}</h1>{{range .Articles}}<article>{{.Title}}</article>{{end}}{{range .Briefs}}{{if .ExternalURL}}<a href="{{.ExternalURL}}">{{.Title}}</a>{{end}}{{end}}</section>{{end}}
		{{define "search"}}<section><h1>{{.Query}}</h1>{{range .Articles}}<article>{{.Title}}</article>{{end}}{{range .Briefs}}{{if .ExternalURL}}<a href="{{.ExternalURL}}">{{.Title}}</a>{{end}}{{end}}</section>{{end}}
	`))

	return &App{
		templates:   tmpl,
		ingestToken: "secret",
		mediaDir:    t.TempDir(),
		ingest:      ingest,
		query:       query,
	}
}

func TestHandleIngestArticleUnauthorized(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/ingest/articles", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	app.handleIngestArticle(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleHealth(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	app.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func TestHandleIngestArticleInvalidPayload(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/ingest/articles", strings.NewReader(`{"external_id":"x"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	app.handleIngestArticle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleIngestArticleCreated(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{
		articleResult: IngestResponse{Created: true, Slug: "deep-dive"},
	}, queryStub{})

	body := map[string]any{
		"external_id":  "ext-1",
		"title":        "Deep Dive",
		"summary":      "Summary",
		"body":         "Body",
		"author":       "Editor",
		"published_at": time.Now().Format(time.RFC3339),
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/ingest/articles", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	app.handleIngestArticle(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if !strings.Contains(rec.Body.String(), `"slug":"deep-dive"`) {
		t.Fatalf("body = %q, want slug", rec.Body.String())
	}
}

func TestHandleIngestArticleIdempotent(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{
		articleResult: IngestResponse{Created: false, Slug: "repeat"},
	}, queryStub{})

	body := map[string]any{
		"external_id":  "ext-1",
		"title":        "Deep Dive",
		"summary":      "Summary",
		"body":         "Body",
		"author":       "Editor",
		"published_at": time.Now().Format(time.RFC3339),
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/ingest/articles", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	app.handleIngestArticle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandleHomeRendersSections(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		home: content.HomePageData{
			HeroArticle: content.Article{Title: "Main Story", Slug: "main-story"},
			Sections: []content.HomeSection{
				{Label: "海外AI公司动态"},
				{Label: "国内AI公司动态"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.handleHome(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Main Story") {
		t.Fatalf("body = %q, want hero title", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "海外AI公司动态") {
		t.Fatalf("body = %q, want home section label", rec.Body.String())
	}
}

func TestHandleHomeRendersBriefExternalLink(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		home: content.HomePageData{
			HeroArticle: content.Article{Title: "Main Story", Slug: "main-story"},
			Sections: []content.HomeSection{
				{
					Label: "海外AI公司动态",
					Briefs: []content.Brief{
						{Title: "OpenAI Update", ExternalURL: "https://example.com/openai"},
					},
				},
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.handleHome(rec, req)

	if !strings.Contains(rec.Body.String(), `href="https://example.com/openai"`) {
		t.Fatalf("body = %q, want brief external link", rec.Body.String())
	}
}

func TestHandleArticleRendersContent(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		article: content.Article{Title: "LLM Infra", Body: "Detailed body"},
	})
	req := httptest.NewRequest(http.MethodGet, "/articles/llm-infra", nil)
	rec := httptest.NewRecorder()

	app.handleArticle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Detailed body") {
		t.Fatalf("body = %q, want article body", rec.Body.String())
	}
}

func TestHandleHomeRendersHeroCover(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		home: content.HomePageData{
			HeroArticle: content.Article{
				Title:      "Main Story",
				Slug:       "main-story",
				Summary:    "Summary",
				CoverImage: "/media/hero.png",
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.handleHome(rec, req)

	if !strings.Contains(rec.Body.String(), `/media/hero.png`) {
		t.Fatalf("body = %q, want hero cover", rec.Body.String())
	}
}

func TestHandleArticleRendersCoverImage(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		article: content.Article{
			Title:      "LLM Infra",
			Body:       "Detailed body",
			CoverImage: "/media/story.png",
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/articles/llm-infra", nil)
	rec := httptest.NewRecorder()

	app.handleArticle(rec, req)

	if !strings.Contains(rec.Body.String(), `/media/story.png`) {
		t.Fatalf("body = %q, want article cover", rec.Body.String())
	}
}

func TestHandleSearchRendersBriefExternalLink(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		search: SearchPageData{
			Query: "Anthropic",
			Briefs: []content.Brief{
				{Title: "Anthropic Update", ExternalURL: "https://example.com/anthropic"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/search?q=Anthropic", nil)
	rec := httptest.NewRecorder()

	app.handleSearch(rec, req)

	if !strings.Contains(rec.Body.String(), `href="https://example.com/anthropic"`) {
		t.Fatalf("body = %q, want brief external link", rec.Body.String())
	}
}

func TestHandleBriefsRendersExternalLink(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, ingestStub{}, queryStub{
		briefs: BriefPageData{
			Briefs: []content.Brief{
				{Title: "Meta update", ExternalURL: "https://example.com/meta"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/briefs", nil)
	rec := httptest.NewRecorder()

	app.handleBriefs(rec, req)

	if !strings.Contains(rec.Body.String(), "https://example.com/meta") {
		t.Fatalf("body = %q, want external link", rec.Body.String())
	}
}

func TestRoutesServeMediaFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cover.txt")
	if err := os.WriteFile(path, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	app := newTestApp(t, ingestStub{}, queryStub{})
	app.mediaDir = dir

	req := httptest.NewRequest(http.MethodGet, "/media/cover.txt", nil)
	rec := httptest.NewRecorder()

	app.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "cover" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "cover")
	}
}
