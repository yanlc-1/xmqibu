package site

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitlab.meitu.com/xmqibu/internal/content"
)

type repoStub struct {
	article          content.Article
	brief            content.Brief
	existing         bool
	homeArticles     []content.Article
	homeBriefs       []content.Brief
	homeTags         []content.Tag
	savedArticle     content.IngestArticleInput
	savedBrief       content.IngestBriefInput
	savedArticleTags []string
	savedBriefTags   []string
}

func (r *repoStub) HasIngestRecord(ctx context.Context, kind, externalID string) (bool, string, error) {
	if r.existing {
		return true, "existing-slug", nil
	}
	return false, "", nil
}

func (r *repoStub) SaveArticle(ctx context.Context, input content.IngestArticleInput) (content.Article, error) {
	r.savedArticle = input
	r.savedArticleTags = append([]string(nil), input.Tags...)
	return content.Article{Slug: input.Slug, Title: input.Title}, nil
}

func (r *repoStub) SaveBrief(ctx context.Context, input content.IngestBriefInput) (content.Brief, error) {
	r.savedBrief = input
	r.savedBriefTags = append([]string(nil), input.Tags...)
	return content.Brief{Slug: input.Slug, Title: input.Title}, nil
}

func (r *repoStub) ListArticles(ctx context.Context, limit int) ([]content.Article, error) {
	return r.homeArticles, nil
}

func (r *repoStub) ListBriefs(ctx context.Context, page, pageSize int) ([]content.Brief, bool, error) {
	return r.homeBriefs, false, nil
}

func (r *repoStub) ListTopTags(ctx context.Context, limit int) ([]content.Tag, error) {
	return r.homeTags, nil
}

func (r *repoStub) GetArticleBySlug(ctx context.Context, slug string) (content.Article, error) {
	if r.article.Slug == "" {
		return content.Article{}, errors.New("not found")
	}
	return r.article, nil
}

func (r *repoStub) GetTagBySlug(ctx context.Context, slug string) (content.Tag, error) {
	return content.Tag{Name: "AI", Slug: slug}, nil
}

func (r *repoStub) ListArticlesByTag(ctx context.Context, slug string, limit int) ([]content.Article, error) {
	return r.homeArticles, nil
}

func (r *repoStub) ListBriefsByTag(ctx context.Context, slug string, page, pageSize int) ([]content.Brief, bool, error) {
	return r.homeBriefs, false, nil
}

func (r *repoStub) Search(ctx context.Context, term string, page, pageSize int) ([]content.Article, []content.Brief, []content.Tag, bool, error) {
	return r.homeArticles, r.homeBriefs, r.homeTags, false, nil
}

func TestIngestArticleReturnsExistingSlug(t *testing.T) {
	t.Parallel()

	repo := &repoStub{existing: true}
	svc := NewService(repo)

	resp, err := svc.IngestArticle(context.Background(), content.IngestArticleInput{
		ExternalID:  "ext-1",
		Title:       "AI infra",
		Summary:     "summary",
		Body:        "body",
		Author:      "team",
		PublishedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("IngestArticle() error = %v", err)
	}
	if resp.Created {
		t.Fatal("IngestArticle() created = true, want false")
	}
	if resp.Slug != "existing-slug" {
		t.Fatalf("IngestArticle() slug = %q, want existing-slug", resp.Slug)
	}
}

func TestIngestArticleNormalizesSlugAndTags(t *testing.T) {
	t.Parallel()

	repo := &repoStub{}
	svc := NewService(repo)

	resp, err := svc.IngestArticle(context.Background(), content.IngestArticleInput{
		ExternalID:  "ext-1",
		Title:       "AI Agents 与未来",
		Summary:     "summary",
		Body:        "body",
		Author:      "team",
		Tags:        []string{" AI ", "数据库", "ai"},
		PublishedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("IngestArticle() error = %v", err)
	}
	if !resp.Created {
		t.Fatal("IngestArticle() created = false, want true")
	}
	if repo.savedArticle.Slug != "ai-agents" {
		t.Fatalf("saved slug = %q, want ai-agents", repo.savedArticle.Slug)
	}
	if len(repo.savedArticleTags) != 2 || repo.savedArticleTags[0] != "ai" || repo.savedArticleTags[1] != "数据库" {
		t.Fatalf("saved tags = %#v, want normalized tags", repo.savedArticleTags)
	}
}

func TestHomePageUsesCuratedData(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &repoStub{
		homeArticles: []content.Article{
			{Title: "Featured", Slug: "featured", Featured: true, PublishedAt: now.Add(-time.Hour)},
			{Title: "Latest", Slug: "latest", PublishedAt: now},
		},
		homeBriefs: []content.Brief{
			{Title: "B1", Slug: "b1", PublishedAt: now},
		},
		homeTags: []content.Tag{
			{Name: "AI", Slug: "ai", UsageCount: 10},
		},
	}
	svc := NewService(repo)

	page, err := svc.HomePage(context.Background())
	if err != nil {
		t.Fatalf("HomePage() error = %v", err)
	}
	if page.HeroArticle.Slug != "featured" {
		t.Fatalf("HeroArticle = %q, want featured", page.HeroArticle.Slug)
	}
}
