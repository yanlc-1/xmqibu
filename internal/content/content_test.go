package content

import (
	"strings"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	t.Parallel()

	got := Slugify("AI Agents 与未来 Systems 2026!")
	want := "ai-agents-systems-2026"
	if got != want {
		t.Fatalf("Slugify() = %q, want %q", got, want)
	}
}

func TestNormalizeTags(t *testing.T) {
	t.Parallel()

	got := NormalizeTags([]string{" AI ", "数据库", "ai", "", "数据库 "})
	want := []string{"ai", "数据库"}
	if len(got) != len(want) {
		t.Fatalf("NormalizeTags() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NormalizeTags()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidateArticleInput(t *testing.T) {
	t.Parallel()

	input := IngestArticleInput{
		ExternalID:  "source-1",
		Title:       "Next-Gen Inference",
		Summary:     "  ",
		Body:        "Deep dive",
		Author:      "Team",
		PublishedAt: time.Now(),
	}

	err := ValidateArticleInput(input)
	if err == nil {
		t.Fatal("ValidateArticleInput() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "summary") {
		t.Fatalf("ValidateArticleInput() error = %q, want mention summary", err)
	}
}

func TestValidateBriefInput(t *testing.T) {
	t.Parallel()

	input := IngestBriefInput{
		ExternalID:  "brief-1",
		Title:       "New release",
		Summary:     "Highlights",
		PublishedAt: time.Time{},
	}

	err := ValidateBriefInput(input)
	if err == nil {
		t.Fatal("ValidateBriefInput() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "published_at") {
		t.Fatalf("ValidateBriefInput() error = %q, want mention published_at", err)
	}
}

func TestBuildHomePageData(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	articles := []Article{
		{
			Title:       "Featured",
			Slug:        "featured",
			Featured:    true,
			PublishedAt: base.Add(-2 * time.Hour),
			Tags:        []Tag{{Name: "海外AI公司动态", Slug: "overseas-ai"}},
		},
		{
			Title:       "GitHub Rising",
			Slug:        "github-rising",
			Featured:    false,
			PublishedAt: base,
			Tags:        []Tag{{Name: "GitHub近3天Star增长Top5", Slug: "github-top"}},
		},
		{
			Title:       "Product Release",
			Slug:        "product-release",
			PublishedAt: base.Add(-24 * time.Hour),
			Tags:        []Tag{{Name: "AI应用与产品动态", Slug: "ai-products"}},
		},
	}
	briefs := []Brief{
		{
			Title:       "Overseas Brief",
			Slug:        "brief-1",
			PublishedAt: base.Add(-time.Hour),
			Tags:        []Tag{{Name: "海外AI公司动态", Slug: "overseas-ai"}},
		},
		{
			Title:       "China Brief",
			Slug:        "brief-2",
			PublishedAt: base.Add(-2 * time.Hour),
			Tags:        []Tag{{Name: "国内AI公司动态", Slug: "china-ai"}},
		},
		{
			Title:       "Product Brief",
			Slug:        "brief-3",
			PublishedAt: base.Add(-3 * time.Hour),
			Tags:        []Tag{{Name: "AI应用与产品动态", Slug: "ai-products"}},
		},
	}
	tags := []Tag{
		{Name: "海外AI公司动态", Slug: "overseas-ai", UsageCount: 12},
		{Name: "国内AI公司动态", Slug: "china-ai", UsageCount: 8},
		{Name: "GitHub近3天Star增长Top5", Slug: "github-top", UsageCount: 9},
		{Name: "AI应用与产品动态", Slug: "ai-products", UsageCount: 7},
	}

	page := BuildHomePageData(articles, briefs, tags)

	if page.HeroArticle.Slug != "featured" {
		t.Fatalf("HeroArticle = %q, want featured article", page.HeroArticle.Slug)
	}
	if len(page.FeaturedArticles) != 2 {
		t.Fatalf("FeaturedArticles len = %d, want 2", len(page.FeaturedArticles))
	}
	if page.FeaturedArticles[0].Slug != "github-rising" {
		t.Fatalf("FeaturedArticles[0] = %q, want github-rising", page.FeaturedArticles[0].Slug)
	}
	if len(page.LatestBriefs) != 3 || page.LatestBriefs[0].Slug != "brief-1" {
		t.Fatalf("LatestBriefs ordering incorrect: %#v", page.LatestBriefs)
	}
	if len(page.TopTags) != 4 || page.TopTags[0].Slug != "overseas-ai" {
		t.Fatalf("TopTags ordering incorrect: %#v", page.TopTags)
	}
	if len(page.Sections) != 4 {
		t.Fatalf("Sections len = %d, want 4", len(page.Sections))
	}
	if page.Sections[0].Slug != "overseas-ai" || page.Sections[0].Label != "海外AI公司动态" {
		t.Fatalf("Sections[0] = %#v, want overseas ai section", page.Sections[0])
	}
	if len(page.Sections[0].Briefs) != 1 || page.Sections[0].Briefs[0].Slug != "brief-1" {
		t.Fatalf("overseas section briefs = %#v, want brief-1", page.Sections[0].Briefs)
	}
	if len(page.Sections[1].Briefs) != 1 || page.Sections[1].Briefs[0].Slug != "brief-2" {
		t.Fatalf("china section briefs = %#v, want brief-2", page.Sections[1].Briefs)
	}
	if len(page.Sections[2].Articles) != 1 || page.Sections[2].Articles[0].Slug != "github-rising" {
		t.Fatalf("github section articles = %#v, want github-rising", page.Sections[2].Articles)
	}
	if len(page.Sections[3].Articles) != 1 || page.Sections[3].Articles[0].Slug != "product-release" {
		t.Fatalf("product section articles = %#v, want product-release", page.Sections[3].Articles)
	}
}
