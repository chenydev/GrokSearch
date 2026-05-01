# Configuration Reference

## Go CLI Configuration

`internal/config/config.go` uses this precedence:

1. Environment variables.
2. JSON config file from `--config` or the default user config path.
3. Built-in defaults.

Default config path uses Go `os.UserConfigDir()`:

- Linux: `$XDG_CONFIG_HOME/grok-search/config.json` or `~/.config/grok-search/config.json`.
- macOS: `~/Library/Application Support/grok-search/config.json`.
- Windows: `%AppData%\grok-search\config.json`.

Go CLI supports:

| Variable | Purpose |
| --- | --- |
| `GUDA_API_KEY` / `guda_api_key` | Derives Grok/Tavily/Firecrawl URLs and keys unless explicit variables override them |
| `GUDA_BASE_URL` / `guda_base_url` | GuDa base URL, default `https://code.guda.studio` |
| `GROK_API_URL` | Explicit Grok-compatible API URL |
| `GROK_API_KEY` | Explicit Grok API key |
| `GROK_MODEL` | Explicit model override |
| `TAVILY_ENABLED` | Enables/disables Tavily |
| `TAVILY_API_URL` | Explicit Tavily API URL |
| `TAVILY_API_KEY` | Explicit Tavily API key |
| `FIRECRAWL_API_URL` | Explicit Firecrawl API URL |
| `FIRECRAWL_API_KEY` | Explicit Firecrawl API key |
| `GROK_DEBUG` | Debug flag reserved for CLI behavior |
| `GROK_RETRY_MAX_ATTEMPTS` | Retry attempts before final failure |
| `GROK_RETRY_MULTIPLIER` | Exponential backoff multiplier |
| `GROK_RETRY_MAX_WAIT` | Max retry wait seconds |

`config set-model` persists `grok_model` to the selected config file. `GROK_MODEL` still takes precedence at runtime.

When `grok_api_url` contains `openrouter`, the config layer appends `:online` to the model if it is not already present.
