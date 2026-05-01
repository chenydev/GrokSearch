package tavily

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
	Title   string  `json:"title,omitempty"`
	URL     string  `json:"url,omitempty"`
	Content string  `json:"content,omitempty"`
	Score   float64 `json:"score,omitempty"`
}

func New(cfg config.Config, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: timeout}}
}

func (c *Client) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	if !c.cfg.TavilyEnabled || c.cfg.TavilyAPIKey == "" {
		return nil, errors.New("TAVILY_API_KEY 未配置")
	}
	body := map[string]any{
		"query":               query,
		"max_results":         maxResults,
		"search_depth":        "advanced",
		"include_raw_content": false,
		"include_answer":      false,
	}
	var out struct {
		Results []SearchResult `json:"results"`
	}
	if err := c.post(ctx, "/search", body, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

func (c *Client) Extract(ctx context.Context, targetURL string) (string, error) {
	if !c.cfg.TavilyEnabled || c.cfg.TavilyAPIKey == "" {
		return "", errors.New("TAVILY_API_KEY 未配置")
	}
	body := map[string]any{"urls": []string{targetURL}, "format": "markdown"}
	var out struct {
		Results []struct {
			RawContent string `json:"raw_content"`
		} `json:"results"`
	}
	if err := c.post(ctx, "/extract", body, &out); err != nil {
		return "", err
	}
	if len(out.Results) == 0 || strings.TrimSpace(out.Results[0].RawContent) == "" {
		return "", errors.New("Tavily 返回空内容")
	}
	return out.Results[0].RawContent, nil
}

func (c *Client) Map(ctx context.Context, targetURL, instructions string, maxDepth, maxBreadth, limit, timeoutSeconds int) (json.RawMessage, error) {
	if !c.cfg.TavilyEnabled || c.cfg.TavilyAPIKey == "" {
		return nil, errors.New("TAVILY_API_KEY 未配置")
	}
	body := map[string]any{
		"url":         targetURL,
		"max_depth":   maxDepth,
		"max_breadth": maxBreadth,
		"limit":       limit,
		"timeout":     timeoutSeconds,
	}
	if instructions != "" {
		body["instructions"] = instructions
	}
	var out json.RawMessage
	if err := c.post(ctx, "/map", body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := strings.TrimRight(c.cfg.TavilyAPIURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.TavilyAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("Tavily HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
