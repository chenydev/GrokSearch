package grok

import (
	"bufio"
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

func New(cfg config.Config, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Models(ctx context.Context) ([]string, error) {
	if err := c.cfg.ValidateGrok(); err != nil {
		return nil, err
	}
	url := strings.TrimRight(c.cfg.GrokAPIURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.GrokAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("models request failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(out.Data))
	for _, item := range out.Data {
		if item.ID != "" {
			models = append(models, item.ID)
		}
	}
	return models, nil
}

func (c *Client) Search(ctx context.Context, query, platform, model string) (string, error) {
	if err := c.cfg.ValidateGrok(); err != nil {
		return "", err
	}
	if model == "" {
		model = c.cfg.GrokModel
	}

	user := query
	if needsTimeContext(query) {
		user = localTimeContext() + "\n" + query
	}
	if platform != "" {
		user += "\n\nFocus platform: " + platform
	}

	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": searchPrompt},
			{"role": "user", "content": user},
		},
		"stream": true,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	var result string
	err = withRetry(ctx, c.cfg, func() (*http.Response, error) {
		url := strings.TrimRight(c.cfg.GrokAPIURL, "/") + "/chat/completions"
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.GrokAPIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return resp, statusError(resp)
		}
		result, err = parseStreamingResponse(resp.Body)
		resp.Body.Close()
		return resp, err
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

func parseStreamingResponse(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var content strings.Builder
	var raw strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		raw.WriteString(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			for _, choice := range chunk.Choices {
				content.WriteString(choice.Delta.Content)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if content.Len() > 0 {
		return content.String(), nil
	}

	var full struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(raw.String()), &full); err == nil {
		for _, choice := range full.Choices {
			content.WriteString(choice.Message.Content)
		}
	}
	return content.String(), nil
}

func statusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	resp.Body.Close()
	return &HTTPStatusError{
		StatusCode: resp.StatusCode,
		RetryAfter: resp.Header.Get("Retry-After"),
		Body:       strings.TrimSpace(string(body)),
	}
}

type HTTPStatusError struct {
	StatusCode int
	RetryAfter string
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

func IsRetryable(err error) bool {
	var status *HTTPStatusError
	if errors.As(err, &status) {
		switch status.StatusCode {
		case 408, 429, 500, 502, 503, 504:
			return true
		default:
			return false
		}
	}
	return true
}
