package web

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"gitlab.meitu.com/xmqibu/internal/content"
)

type IngestResponse struct {
	Created bool   `json:"created"`
	Slug    string `json:"slug"`
}

type BriefPageData struct {
	Briefs   []content.Brief
	Page     int
	PageSize int
	HasMore  bool
}

type TagPageData struct {
	Tag      content.Tag
	Articles []content.Article
	Briefs   []content.Brief
	Page     int
	PageSize int
	HasMore  bool
}

type SearchPageData struct {
	Query    string
	Articles []content.Article
	Briefs   []content.Brief
	Tags     []content.Tag
	Page     int
	PageSize int
	HasMore  bool
}

type IngestService interface {
	IngestArticle(ctx context.Context, input content.IngestArticleInput) (IngestResponse, error)
	IngestBrief(ctx context.Context, input content.IngestBriefInput) (IngestResponse, error)
}

type QueryService interface {
	HomePage(ctx context.Context) (content.HomePageData, error)
	ArticleBySlug(ctx context.Context, slug string) (content.Article, error)
	Briefs(ctx context.Context, page, pageSize int) (BriefPageData, error)
	TagPage(ctx context.Context, slug string, page, pageSize int) (TagPageData, error)
	Search(ctx context.Context, term string, page, pageSize int) (SearchPageData, error)
}

type App struct {
	templates   *template.Template
	ingestToken string
	mediaDir    string
	ingest      IngestService
	query       QueryService
}

func NewApp(templates *template.Template, ingestToken string, mediaDir string, ingest IngestService, query QueryService) *App {
	return &App{
		templates:   templates,
		ingestToken: ingestToken,
		mediaDir:    mediaDir,
		ingest:      ingest,
		query:       query,
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	staticFS, err := fs.Sub(StaticFS(), ".")
	if err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}
	if strings.TrimSpace(a.mediaDir) != "" {
		mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(a.mediaDir))))
	}
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/", a.handleHome)
	mux.HandleFunc("/articles/", a.handleArticle)
	mux.HandleFunc("/briefs", a.handleBriefs)
	mux.HandleFunc("/tags/", a.handleTag)
	mux.HandleFunc("/search", a.handleSearch)
	mux.HandleFunc("/api/ingest/articles", a.handleIngestArticle)
	mux.HandleFunc("/api/ingest/briefs", a.handleIngestBrief)
	return mux
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := a.query.HomePage(r.Context())
	if err != nil {
		http.Error(w, "failed to load homepage", http.StatusInternalServerError)
		return
	}
	a.render(w, "home", data)
}

func (a *App) handleArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/articles/")
	if slug == "" || slug == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	article, err := a.query.ArticleBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "article not found", http.StatusNotFound)
		return
	}
	a.render(w, "article", struct {
		Article content.Article
	}{Article: article})
}

func (a *App) handleBriefs(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r, "page")
	data, err := a.query.Briefs(r.Context(), page, 10)
	if err != nil {
		http.Error(w, "failed to load briefs", http.StatusInternalServerError)
		return
	}
	a.render(w, "briefs", data)
}

func (a *App) handleTag(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/tags/")
	if slug == "" || slug == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	page := parsePage(r, "page")
	data, err := a.query.TagPage(r.Context(), slug, page, 10)
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}
	a.render(w, "tag", data)
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r, "page")
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	data, err := a.query.Search(r.Context(), query, page, 10)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	a.render(w, "search", data)
}

func (a *App) handleIngestArticle(w http.ResponseWriter, r *http.Request) {
	if !a.authorize(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input content.IngestArticleInput
	if err := decodeJSON(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := content.ValidateArticleInput(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.ingest.IngestArticle(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to ingest article", http.StatusInternalServerError)
		return
	}
	writeJSON(w, statusForCreated(resp.Created), resp)
}

func (a *App) handleIngestBrief(w http.ResponseWriter, r *http.Request) {
	if !a.authorize(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input content.IngestBriefInput
	if err := decodeJSON(r, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := content.ValidateBriefInput(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.ingest.IngestBrief(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to ingest brief", http.StatusInternalServerError)
		return
	}
	writeJSON(w, statusForCreated(resp.Created), resp)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}

func (a *App) authorize(r *http.Request) bool {
	expected := strings.TrimSpace(a.ingestToken)
	if expected == "" {
		return false
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	return auth == "Bearer "+expected
}

func decodeJSON(r *http.Request, dest any) error {
	if r.Method != http.MethodPost {
		return errors.New("method not allowed")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func statusForCreated(created bool) int {
	if created {
		return http.StatusCreated
	}
	return http.StatusOK
}

func parsePage(r *http.Request, key string) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 1
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 {
		return 1
	}
	return page
}
