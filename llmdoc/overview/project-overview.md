# Project Overview

## Product Purpose

Grok Search provides a local Go CLI for web search, URL fetch, and site mapping. It is intended for direct terminal use, scripting, and AI-client skill integration with lower token overhead than MCP tool schemas.

It combines:

- Grok-compatible OpenAI-style chat/completions for AI-generated search answers and model listing.
- Tavily Search, Extract, and Map for source supplementation, page extraction, and site mapping.
- Firecrawl Search/Scrape for source supplementation and fetch fallback when configured.

## Package Layout

```text
cmd/grok-search/
  main.go
internal/
  cli/
  config/
  firecrawl/
  grok/
  sources/
  tavily/
```

## Public Entry Points

- Console binary: `grok-search`.
- Main package: `./cmd/grok-search`.
- Build command: `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search`.

## Documentation Layout

- `README.md`: Go CLI install, configuration, and command documentation.
- `.github/workflows/build.yml`: test and cross-platform release artifact workflow.
- `llmdoc/`: internal stable project knowledge for future coding work.
