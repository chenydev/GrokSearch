# Go CLI Flow

## Entry Point

`cmd/grok-search/main.go` constructs the Cobra root command from `internal/cli.NewRootCommand()` and exits non-zero on errors.

## Command Surface

`internal/cli/root.go` wires:

- `search`
- `fetch`
- `map`
- `config get`
- `config check`
- `config set-model`
- `models`
- `last`

The CLI intentionally does not implement MCP `get_sources` as a separate session command. Sources are returned inline with `--format json`, written by `--sources`, or cached as the latest search result for `last`.

## Runtime Flow

### Search

1. Load config from file plus environment.
2. Apply `--model` if provided.
3. Call Grok `/chat/completions` through `internal/grok`.
4. Split answer and sources through `internal/sources`.
5. Optionally collect `--extra-sources` through Tavily and/or Firecrawl.
6. Print text or JSON.
7. Write latest search result to user cache.

### Fetch

1. Try Tavily Extract.
2. If Tavily fails or is not configured, try Firecrawl Scrape.
3. Print Markdown or JSON with provider metadata.

### Map

1. Call Tavily Map.
2. Print raw JSON or pretty JSON.

## Dependency Policy

Go CLI v1 uses Cobra for subcommands and help text. HTTP, JSON, config, retry, SSE parsing, output formatting, and source handling are implemented with the Go standard library.
