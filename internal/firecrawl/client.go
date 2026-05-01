package firecrawl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GuDaStudio/GrokSearch/internal/config"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

type SearchResult struct {
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

func New(cfg config.Config, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: timeout}}
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if c.cfg.FirecrawlAPIKey == "" {
		return nil, errors.New("FIRECRAWL_API_KEY 未配置")
	}
	body := map[string]any{"query": query, "limit": limit}
	var out struct {
		Data struct {
			Web []SearchResult `json:"web"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/search", body, &out); err != nil {
		return nil, err
	}
	return out.Data.Web, nil
}

func (c *Client) Scrape(ctx context.Context, targetURL string, attempts int) (string, error) {
	if c.cfg.FirecrawlAPIKey == "" {
		return "", errors.New("FIRECRAWL_API_KEY 未配置")
	}
	if attempts <= 0 {
		attempts = 3
	}
	var last error
	for i := 0; i < attempts; i++ {
		body := map[string]any{
			"url":     targetURL,
			"formats": []string{"markdown"},
			"timeout": 60000,
			"waitFor": (i + 1) * 1500,
		}
		var out struct {
			Data struct {
				Markdown string `json:"markdown"`
			} `json:"data"`
		}
		if err := c.post(ctx, "/scrape", body, &out); err != nil {
			return "", err
		}
		if strings.TrimSpace(out.Data.Markdown) != "" {
			return out.Data.Markdown, nil
		}
		last = errors.New("Firecrawl 返回空 markdown")
	}
	return "", last
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := strings.TrimRight(c.cfg.FirecrawlAPIURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.FirecrawlAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("Firecrawl HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
