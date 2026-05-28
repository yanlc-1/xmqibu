package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.meitu.com/xmqibu/internal/content"
	"gitlab.meitu.com/xmqibu/internal/web"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}
}

func (c *Client) PushArticle(ctx context.Context, input content.IngestArticleInput) (web.IngestResponse, error) {
	return c.post(ctx, "/api/ingest/articles", input)
}

func (c *Client) PushBrief(ctx context.Context, input content.IngestBriefInput) (web.IngestResponse, error) {
	return c.post(ctx, "/api/ingest/briefs", input)
}

func (c *Client) post(ctx context.Context, path string, payload any) (web.IngestResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return web.IngestResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return web.IngestResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return web.IngestResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return web.IngestResponse{}, fmt.Errorf("ingest request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	var result web.IngestResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return web.IngestResponse{}, err
	}
	return result, nil
}
