package site

import (
	"context"

	"gitlab.meitu.com/xmqibu/internal/content"
	"gitlab.meitu.com/xmqibu/internal/web"
)

type Repository interface {
	HasIngestRecord(ctx context.Context, kind, externalID string) (bool, string, error)
	SaveArticle(ctx context.Context, input content.IngestArticleInput) (content.Article, error)
	SaveBrief(ctx context.Context, input content.IngestBriefInput) (content.Brief, error)
	ListArticles(ctx context.Context, limit int) ([]content.Article, error)
	ListBriefs(ctx context.Context, page, pageSize int) ([]content.Brief, bool, error)
	ListTopTags(ctx context.Context, limit int) ([]content.Tag, error)
	GetArticleBySlug(ctx context.Context, slug string) (content.Article, error)
	GetTagBySlug(ctx context.Context, slug string) (content.Tag, error)
	ListArticlesByTag(ctx context.Context, slug string, limit int) ([]content.Article, error)
	ListBriefsByTag(ctx context.Context, slug string, page, pageSize int) ([]content.Brief, bool, error)
	Search(ctx context.Context, term string, page, pageSize int) ([]content.Article, []content.Brief, []content.Tag, bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IngestArticle(ctx context.Context, input content.IngestArticleInput) (web.IngestResponse, error) {
	exists, slug, err := s.repo.HasIngestRecord(ctx, "article", input.ExternalID)
	if err != nil {
		return web.IngestResponse{}, err
	}
	if exists {
		return web.IngestResponse{Created: false, Slug: slug}, nil
	}

	input.Tags = content.NormalizeTags(input.Tags)
	if input.Slug == "" {
		input.Slug = content.Slugify(input.Title)
	}

	article, err := s.repo.SaveArticle(ctx, input)
	if err != nil {
		return web.IngestResponse{}, err
	}
	return web.IngestResponse{Created: true, Slug: article.Slug}, nil
}

func (s *Service) IngestBrief(ctx context.Context, input content.IngestBriefInput) (web.IngestResponse, error) {
	exists, slug, err := s.repo.HasIngestRecord(ctx, "brief", input.ExternalID)
	if err != nil {
		return web.IngestResponse{}, err
	}
	if exists {
		return web.IngestResponse{Created: false, Slug: slug}, nil
	}

	input.Tags = content.NormalizeTags(input.Tags)
	if input.Slug == "" {
		input.Slug = content.Slugify(input.Title)
	}

	brief, err := s.repo.SaveBrief(ctx, input)
	if err != nil {
		return web.IngestResponse{}, err
	}
	return web.IngestResponse{Created: true, Slug: brief.Slug}, nil
}

func (s *Service) HomePage(ctx context.Context) (content.HomePageData, error) {
	articles, err := s.repo.ListArticles(ctx, 6)
	if err != nil {
		return content.HomePageData{}, err
	}
	briefs, _, err := s.repo.ListBriefs(ctx, 1, 6)
	if err != nil {
		return content.HomePageData{}, err
	}
	tags, err := s.repo.ListTopTags(ctx, 8)
	if err != nil {
		return content.HomePageData{}, err
	}
	return content.BuildHomePageData(articles, briefs, tags), nil
}

func (s *Service) ArticleBySlug(ctx context.Context, slug string) (content.Article, error) {
	return s.repo.GetArticleBySlug(ctx, slug)
}

func (s *Service) Briefs(ctx context.Context, page, pageSize int) (web.BriefPageData, error) {
	briefs, hasMore, err := s.repo.ListBriefs(ctx, page, pageSize)
	if err != nil {
		return web.BriefPageData{}, err
	}
	return web.BriefPageData{Briefs: briefs, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

func (s *Service) TagPage(ctx context.Context, slug string, page, pageSize int) (web.TagPageData, error) {
	tag, err := s.repo.GetTagBySlug(ctx, slug)
	if err != nil {
		return web.TagPageData{}, err
	}
	articles, err := s.repo.ListArticlesByTag(ctx, slug, 6)
	if err != nil {
		return web.TagPageData{}, err
	}
	briefs, hasMore, err := s.repo.ListBriefsByTag(ctx, slug, page, pageSize)
	if err != nil {
		return web.TagPageData{}, err
	}
	return web.TagPageData{
		Tag:      tag,
		Articles: articles,
		Briefs:   briefs,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

func (s *Service) Search(ctx context.Context, term string, page, pageSize int) (web.SearchPageData, error) {
	articles, briefs, tags, hasMore, err := s.repo.Search(ctx, term, page, pageSize)
	if err != nil {
		return web.SearchPageData{}, err
	}
	return web.SearchPageData{
		Query:    term,
		Articles: articles,
		Briefs:   briefs,
		Tags:     tags,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}
