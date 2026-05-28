package ingest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.meitu.com/xmqibu/internal/content"
)

func TestClientPushArticleSendsTokenAndPayload(t *testing.T) {
	t.Parallel()

	var (
		gotAuth  string
		gotInput content.IngestArticleInput
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ingest/articles" {
			t.Fatalf("path = %q, want /api/ingest/articles", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotInput); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"created": true,
			"slug":    "agent-systems-in-2026",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", server.Client())
	resp, err := client.PushArticle(context.Background(), content.IngestArticleInput{
		ExternalID:  "demo-1",
		Title:       "Agent Systems in 2026",
		Summary:     "summary",
		Body:        "body",
		Author:      "Codex",
		Tags:        []string{"AI"},
		PublishedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("PushArticle() error = %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", gotAuth)
	}
	if gotInput.ExternalID != "demo-1" || gotInput.Title != "Agent Systems in 2026" {
		t.Fatalf("payload = %#v, want article payload", gotInput)
	}
	if !resp.Created || resp.Slug != "agent-systems-in-2026" {
		t.Fatalf("response = %#v, want created slug", resp)
	}
}

func TestClientPushBriefReturnsErrorOnNonSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token", server.Client())
	_, err := client.PushBrief(context.Background(), content.IngestBriefInput{
		ExternalID:  "brief-1",
		Title:       "Brief",
		Summary:     "summary",
		PublishedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("PushBrief() error = nil, want error")
	}
}
