package content

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	StatusPublished = "published"
	StatusDraft     = "draft"
)

type Article struct {
	ID          int64
	Title       string
	Slug        string
	Summary     string
	CoverImage  string
	Body        string
	Author      string
	SourceLinks []string
	PublishedAt time.Time
	Featured    bool
	Status      string
	Tags        []Tag
}

type Brief struct {
	ID          int64
	Title       string
	Slug        string
	Summary     string
	ExternalURL string
	Importance  int
	SourceLinks []string
	PublishedAt time.Time
	Status      string
	Tags        []Tag
}

type Tag struct {
	ID         int64
	Name       string
	Slug       string
	UsageCount int
}

type IngestArticleInput struct {
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	CoverImage  string    `json:"cover_image"`
	Body        string    `json:"body"`
	Author      string    `json:"author"`
	SourceLinks []string  `json:"source_links"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"published_at"`
	Featured    bool      `json:"featured"`
}

type IngestBriefInput struct {
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	ExternalURL string    `json:"external_url"`
	Importance  int       `json:"importance"`
	SourceLinks []string  `json:"source_links"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"published_at"`
}

type HomePageData struct {
	HeroArticle      Article
	FeaturedArticles []Article
	LatestBriefs     []Brief
	TopTags          []Tag
	Sections         []HomeSection
}

type HomeSection struct {
	Slug        string
	Label       string
	Description string
	Articles    []Article
	Briefs      []Brief
}

type sectionDefinition struct {
	Slug        string
	Label       string
	Description string
}

var fixedHomeSections = []sectionDefinition{
	{Slug: "overseas-ai", Label: "海外AI公司动态", Description: "跟踪海外 AI 公司发布、合作、融资与战略变化。"},
	{Slug: "china-ai", Label: "国内AI公司动态", Description: "聚合国内 AI 公司在模型、产品和组织层面的最新动作。"},
	{Slug: "github-top", Label: "GitHub 近3天 Star 增长 Top5", Description: "筛选近三天热度增长最快的开源项目与工具。"},
	{Slug: "ai-products", Label: "AI应用与产品动态", Description: "覆盖 AI 应用发布、产品更新和关键功能迭代。"},
}

var slugPattern = regexp.MustCompile(`-+`)

func Slugify(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case !lastDash:
			builder.WriteRune('-')
			lastDash = true
		}
	}

	slug := strings.Trim(builder.String(), "-")
	slug = slugPattern.ReplaceAllString(slug, "-")
	return slug
}

func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		normalized = append(normalized, cleaned)
	}
	sort.Strings(normalized)
	return normalized
}

func ValidateArticleInput(input IngestArticleInput) error {
	if strings.TrimSpace(input.ExternalID) == "" {
		return errors.New("external_id is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(input.Summary) == "" {
		return errors.New("summary is required")
	}
	if strings.TrimSpace(input.Body) == "" {
		return errors.New("body is required")
	}
	if strings.TrimSpace(input.Author) == "" {
		return errors.New("author is required")
	}
	if input.PublishedAt.IsZero() {
		return errors.New("published_at is required")
	}
	return nil
}

func ValidateBriefInput(input IngestBriefInput) error {
	if strings.TrimSpace(input.ExternalID) == "" {
		return errors.New("external_id is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(input.Summary) == "" {
		return errors.New("summary is required")
	}
	if input.PublishedAt.IsZero() {
		return errors.New("published_at is required")
	}
	return nil
}

func BuildHomePageData(articles []Article, briefs []Brief, tags []Tag) HomePageData {
	sortedArticles := append([]Article(nil), articles...)
	sort.Slice(sortedArticles, func(i, j int) bool {
		if sortedArticles[i].Featured != sortedArticles[j].Featured {
			return sortedArticles[i].Featured
		}
		return sortedArticles[i].PublishedAt.After(sortedArticles[j].PublishedAt)
	})

	var hero Article
	if len(sortedArticles) > 0 {
		hero = sortedArticles[0]
	}

	curated := make([]Article, 0, 2)
	for _, article := range sortedArticles {
		if hero.Slug != "" && article.Slug == hero.Slug {
			continue
		}
		curated = append(curated, article)
		if len(curated) == 2 {
			break
		}
	}

	sortedBriefs := append([]Brief(nil), briefs...)
	sort.Slice(sortedBriefs, func(i, j int) bool {
		return sortedBriefs[i].PublishedAt.After(sortedBriefs[j].PublishedAt)
	})

	sortedTags := append([]Tag(nil), tags...)
	sort.Slice(sortedTags, func(i, j int) bool {
		if sortedTags[i].UsageCount != sortedTags[j].UsageCount {
			return sortedTags[i].UsageCount > sortedTags[j].UsageCount
		}
		return sortedTags[i].Name < sortedTags[j].Name
	})

	return HomePageData{
		HeroArticle:      hero,
		FeaturedArticles: curated,
		LatestBriefs:     sortedBriefs,
		TopTags:          sortedTags,
		Sections:         buildHomeSections(sortedArticles, sortedBriefs),
	}
}

func buildHomeSections(articles []Article, briefs []Brief) []HomeSection {
	sections := make([]HomeSection, 0, len(fixedHomeSections))
	for _, definition := range fixedHomeSections {
		section := HomeSection{
			Slug:        definition.Slug,
			Label:       definition.Label,
			Description: definition.Description,
		}
		for _, article := range articles {
			if hasTagSlug(article.Tags, definition.Slug) {
				section.Articles = append(section.Articles, article)
			}
		}
		for _, brief := range briefs {
			if hasTagSlug(brief.Tags, definition.Slug) {
				section.Briefs = append(section.Briefs, brief)
			}
		}
		sections = append(sections, section)
	}
	return sections
}

func hasTagSlug(tags []Tag, slug string) bool {
	for _, tag := range tags {
		if tag.Slug == slug {
			return true
		}
	}
	return false
}
