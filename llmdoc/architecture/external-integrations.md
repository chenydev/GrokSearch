# External Integrations

## Grok-Compatible API

`internal/grok.Client` uses an OpenAI-compatible API URL and sends requests to:

```text
{GROK_API_URL}/chat/completions
```

Model discovery and connection checks call:

```text
{GROK_API_URL}/models
```

Streaming parsing supports `data:` SSE chunks and a fallback full JSON body with `choices[0].message.content`.

## Retry Behavior

`internal/grok` retries selected HTTP errors. Retryable HTTP status codes are:

```text
408, 429, 500, 502, 503, 504
```

For `429`, `Retry-After` is parsed as either seconds or an HTTP date. Retry knobs come from config/env values `GROK_RETRY_MAX_ATTEMPTS`, `GROK_RETRY_MULTIPLIER`, and `GROK_RETRY_MAX_WAIT`.

## Tavily

Tavily is used in `internal/tavily` for:

- `/extract`: `web_fetch` primary extraction path.
- `/search`: optional `web_search(extra_sources)` source supplementation.
- `/map`: `web_map`.

`web_map` requires `TAVILY_API_KEY`; `web_fetch` can still succeed through Firecrawl when Tavily is missing or fails.

## Firecrawl

Firecrawl is used in `internal/firecrawl` for:

- `/search`: optional `web_search(extra_sources)` source supplementation.
- `/scrape`: `web_fetch` fallback path.

Firecrawl scrape retries empty Markdown responses by increasing `waitFor`.

## Logging

Go CLI v1 keeps logging minimal and writes command errors to stderr.
