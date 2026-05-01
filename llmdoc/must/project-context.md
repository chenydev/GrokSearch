# Project Context

## Identity

Grok Search is a Go CLI that provides web search, URL fetch, and site mapping through Grok-compatible, Tavily, and Firecrawl APIs.

## Runtime Shape

- Go CLI entry point: `cmd/grok-search/main.go`
- Go module: `github.com/GuDaStudio/GrokSearch`

## Core Ownership Boundaries

- `cmd/grok-search`: Go CLI entry point.
- `internal/cli`: Go CLI command wiring.
- `internal/config`: Go config loading, env precedence, GuDa-derived defaults, and model persistence.
- `internal/grok`: Grok-compatible `/models` and `/chat/completions` client, SSE parsing, retry behavior, and time-context handling.
- `internal/tavily`: Tavily Search, Extract, and Map client.
- `internal/firecrawl`: Firecrawl Search and Scrape client.
- `internal/sources`: answer/source splitting, normalization, merging, and provider conversion.

## Current Scope

The Go CLI exposes search, fetch, map, config, models, and last-result commands. It intentionally does not include MCP, Python packaging, `get_sources` sessions, or the old `plan_*` tool family.
