package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultModel        = "grok-4-fast"
	defaultTavilyURL    = "https://api.tavily.com"
	defaultFirecrawlURL = "https://api.firecrawl.dev/v2"
	defaultGuDaBaseURL  = "https://code.guda.studio"
)

type Config struct {
	GuDaAPIKey       string  `json:"guda_api_key,omitempty"`
	GuDaBaseURL      string  `json:"guda_base_url,omitempty"`
	GrokAPIURL       string  `json:"grok_api_url,omitempty"`
	GrokAPIKey       string  `json:"grok_api_key,omitempty"`
	GrokModel        string  `json:"grok_model,omitempty"`
	TavilyEnabled    bool    `json:"tavily_enabled"`
	TavilyAPIURL     string  `json:"tavily_api_url,omitempty"`
	TavilyAPIKey     string  `json:"tavily_api_key,omitempty"`
	FirecrawlAPIURL  string  `json:"firecrawl_api_url,omitempty"`
	FirecrawlAPIKey  string  `json:"firecrawl_api_key,omitempty"`
	Debug            bool    `json:"debug"`
	RetryMaxAttempts int     `json:"retry_max_attempts"`
	RetryMultiplier  float64 `json:"retry_multiplier"`
	RetryMaxWait     int     `json:"retry_max_wait"`
}

type Loaded struct {
	Config Config
	Path   string
}

func Default() Config {
	return Config{
		GuDaBaseURL:      defaultGuDaBaseURL,
		GrokModel:        defaultModel,
		TavilyEnabled:    true,
		TavilyAPIURL:     defaultTavilyURL,
		FirecrawlAPIURL:  defaultFirecrawlURL,
		RetryMaxAttempts: 3,
		RetryMultiplier:  1,
		RetryMaxWait:     10,
	}
}

func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "grok-search", "config.json")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".config", "grok-search", "config.json")
	}
	return filepath.Join(".grok-search", "config.json")
}

func Load(path string) (Loaded, error) {
	if path == "" {
		path = DefaultPath()
	}

	cfg := Default()
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &cfg); err != nil {
			return Loaded{}, fmt.Errorf("read config %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Loaded{}, fmt.Errorf("read config %s: %w", path, err)
	}

	applyEnv(&cfg)
	applyGuDa(&cfg)
	cfg.GrokModel = applyModelSuffix(cfg.GrokAPIURL, cfg.GrokModel)

	return Loaded{Config: cfg, Path: path}, nil
}

func SaveModel(path, model string) error {
	if path == "" {
		path = DefaultPath()
	}

	data := map[string]any{}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &data)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	data["grok_model"] = model
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func (c Config) ValidateGrok() error {
	if strings.TrimSpace(c.GrokAPIURL) == "" {
		return errors.New("GROK_API_URL 未配置")
	}
	if strings.TrimSpace(c.GrokAPIKey) == "" {
		return errors.New("GROK_API_KEY 未配置")
	}
	return nil
}

func (c Config) RetryMaxWaitDuration() time.Duration {
	if c.RetryMaxWait <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.RetryMaxWait) * time.Second
}

func Mask(key string) string {
	if key == "" {
		return "未配置"
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func applyGuDa(cfg *Config) {
	key := cfg.GuDaAPIKey
	if key == "" {
		return
	}
	base := strings.TrimRight(cfg.GuDaBaseURL, "/")
	if base == "" {
		base = defaultGuDaBaseURL
	}
	if cfg.GrokAPIURL == "" {
		cfg.GrokAPIURL = base + "/grok/v1"
	}
	if cfg.GrokAPIKey == "" {
		cfg.GrokAPIKey = key
	}
	if cfg.TavilyAPIURL == "" || cfg.TavilyAPIURL == defaultTavilyURL {
		cfg.TavilyAPIURL = base + "/tavily"
	}
	if cfg.TavilyAPIKey == "" {
		cfg.TavilyAPIKey = key
	}
	if cfg.FirecrawlAPIURL == "" || cfg.FirecrawlAPIURL == defaultFirecrawlURL {
		cfg.FirecrawlAPIURL = base + "/firecrawl"
	}
	if cfg.FirecrawlAPIKey == "" {
		cfg.FirecrawlAPIKey = key
	}
}

func applyEnv(cfg *Config) {
	setString(&cfg.GuDaAPIKey, "GUDA_API_KEY")
	setString(&cfg.GuDaBaseURL, "GUDA_BASE_URL")
	setString(&cfg.GrokAPIURL, "GROK_API_URL")
	setString(&cfg.GrokAPIKey, "GROK_API_KEY")
	setString(&cfg.GrokModel, "GROK_MODEL")
	setString(&cfg.TavilyAPIURL, "TAVILY_API_URL")
	setString(&cfg.TavilyAPIKey, "TAVILY_API_KEY")
	setString(&cfg.FirecrawlAPIURL, "FIRECRAWL_API_URL")
	setString(&cfg.FirecrawlAPIKey, "FIRECRAWL_API_KEY")
	setBool(&cfg.TavilyEnabled, "TAVILY_ENABLED")
	setBool(&cfg.Debug, "GROK_DEBUG")
	setInt(&cfg.RetryMaxAttempts, "GROK_RETRY_MAX_ATTEMPTS")
	setFloat(&cfg.RetryMultiplier, "GROK_RETRY_MULTIPLIER")
	setInt(&cfg.RetryMaxWait, "GROK_RETRY_MAX_WAIT")
}

func setString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func setBool(dst *bool, key string) {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "on":
			*dst = true
		case "0", "false", "no", "off":
			*dst = false
		}
	}
}

func setInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
}

func setFloat(dst *float64, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			*dst = n
		}
	}
}

func applyModelSuffix(apiURL, model string) string {
	if model == "" {
		model = defaultModel
	}
	if strings.Contains(strings.ToLower(apiURL), "openrouter") && !strings.Contains(model, ":online") {
		return model + ":online"
	}
	return model
}
