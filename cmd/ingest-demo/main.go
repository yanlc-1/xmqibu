package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.meitu.com/xmqibu/internal/content"
	"gitlab.meitu.com/xmqibu/internal/ingest"
)

func main() {
	var (
		mode        = flag.String("mode", "article", "ingest mode: article or brief")
		baseURL     = flag.String("base-url", envOr("SITE_BASE_URL", "http://127.0.0.1:8080"), "site base url")
		token       = flag.String("token", envOr("INGEST_TOKEN", ""), "ingest token")
		externalID  = flag.String("external-id", "", "external content id")
		title       = flag.String("title", "", "content title")
		summary     = flag.String("summary", "", "content summary")
		body        = flag.String("body", "", "article body")
		author      = flag.String("author", "Automation", "article author")
		externalURL = flag.String("external-url", "", "brief external url")
		importance  = flag.Int("importance", 1, "brief importance")
		tags        = flag.String("tags", "AI,Agent", "comma-separated tags")
		sourceLinks = flag.String("source-links", "", "comma-separated source links")
		featured    = flag.Bool("featured", false, "mark article as featured")
		publishedAt = flag.String("published-at", "", "published time in RFC3339, defaults to now")
	)
	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		log.Fatal("missing ingest token: set -token or INGEST_TOKEN")
	}
	if strings.TrimSpace(*externalID) == "" || strings.TrimSpace(*title) == "" || strings.TrimSpace(*summary) == "" {
		log.Fatal("missing required flags: -external-id, -title, -summary")
	}

	when := time.Now()
	if strings.TrimSpace(*publishedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, *publishedAt)
		if err != nil {
			log.Fatalf("invalid -published-at: %v", err)
		}
		when = parsed
	}

	client := ingest.NewClient(*baseURL, *token, &http.Client{Timeout: 15 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "article":
		if strings.TrimSpace(*body) == "" {
			log.Fatal("missing required flag: -body for article mode")
		}
		resp, err := client.PushArticle(ctx, content.IngestArticleInput{
			ExternalID:  *externalID,
			Title:       *title,
			Summary:     *summary,
			Body:        *body,
			Author:      *author,
			Tags:        splitCSV(*tags),
			SourceLinks: splitCSV(*sourceLinks),
			PublishedAt: when,
			Featured:    *featured,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("article pushed: created=%t slug=%s\n", resp.Created, resp.Slug)
	case "brief":
		resp, err := client.PushBrief(ctx, content.IngestBriefInput{
			ExternalID:  *externalID,
			Title:       *title,
			Summary:     *summary,
			ExternalURL: *externalURL,
			Importance:  *importance,
			Tags:        splitCSV(*tags),
			SourceLinks: splitCSV(*sourceLinks),
			PublishedAt: when,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("brief pushed: created=%t slug=%s\n", resp.Created, resp.Slug)
	default:
		log.Fatalf("unsupported mode %q", *mode)
	}
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
