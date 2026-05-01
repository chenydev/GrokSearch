package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrecedenceEnvOverFile(t *testing.T) {
	t.Setenv("GROK_API_URL", "https://env.example/v1")
	t.Setenv("GROK_API_KEY", "env-key")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"grok_api_url":"https://file.example/v1","grok_api_key":"file-key","grok_model":"file-model"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Config.GrokAPIURL != "https://env.example/v1" {
		t.Fatalf("env should override file, got %q", loaded.Config.GrokAPIURL)
	}
	if loaded.Config.GrokAPIKey != "env-key" {
		t.Fatalf("env key should override file")
	}
	if loaded.Config.GrokModel != "file-model" {
		t.Fatalf("file model should remain when env not set")
	}
}

func TestGuDaDerivesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"guda_api_key":"guda-key","guda_base_url":"https://code.example"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := loaded.Config
	if cfg.GrokAPIURL != "https://code.example/grok/v1" || cfg.GrokAPIKey != "guda-key" {
		t.Fatalf("unexpected grok config: %#v", cfg)
	}
	if cfg.TavilyAPIURL != "https://code.example/tavily" || cfg.TavilyAPIKey != "guda-key" {
		t.Fatalf("unexpected tavily config: %#v", cfg)
	}
	if cfg.FirecrawlAPIURL != "https://code.example/firecrawl" || cfg.FirecrawlAPIKey != "guda-key" {
		t.Fatalf("unexpected firecrawl config: %#v", cfg)
	}
}
