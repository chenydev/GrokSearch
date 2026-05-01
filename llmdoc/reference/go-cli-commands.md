# Go CLI Commands

## Global Flags

- `--config`: explicit config file path.
- `--format`: `text` or `json`.
- `--timeout`: request timeout, default `120s`.

## Commands

### `search QUERY`

Flags:

- `--platform`
- `--model`
- `--extra-sources`
- `--sources`
- `--show-sources`

Default text output prints only answer content. JSON output includes `content`, `sources`, `sources_count`, and `model`.

### `fetch URL`

Fetches Markdown through Tavily first, with Firecrawl fallback.

JSON output includes `url`, `provider`, and `content`.

### `map URL`

Calls Tavily Map.

Flags:

- `--instructions`
- `--max-depth`
- `--max-breadth`
- `--limit`
- `--map-timeout`

### `config get`

Prints effective configuration with API keys masked.

### `config check`

Prints effective configuration and tests Grok `/models`.

### `config set-model MODEL`

Persists `grok_model` to the selected config file.

### `models`

Lists Grok-compatible model IDs from `/models`.

### `last`

Reads the latest cached search result from user cache. This is not a long-lived session store.
