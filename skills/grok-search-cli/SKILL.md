---
name: grok-search-cli
description: Use the local Grok Search Go CLI for current web search, URL fetch, site mapping, source-grounded answers, and configuration checks from any agent that can run shell commands.
---

# Grok Search CLI

## Overview

Use this skill when a task needs current web information, source collection, URL content extraction, or Tavily site mapping through the local `grok-search` CLI.

Use English for CLI queries and tool-facing instructions unless the query is inherently Chinese-specific. Keep final user-facing answers in Simplified Chinese.

Resolve the command in this order:

1. Use `grok-search` if it is available on `PATH`.
2. If working inside this repository and `./dist/grok-search` exists, use that.
3. If working inside this repository and no binary exists, build `./dist/grok-search`.

## Command Selection

- Search current web information:

```bash
grok-search search "query" --format json
```

- Fetch one URL as Markdown:

```bash
grok-search fetch "https://example.com/page"
```

- Map a documentation site:

```bash
grok-search map "https://docs.example.com" --instructions "only API docs" --format json
```

- Check configuration:

```bash
grok-search config check
```

If `grok-search` is not on `PATH` but `./dist/grok-search` exists, replace `grok-search` with `./dist/grok-search`.

If no binary exists and the workspace is this repository, build it:

```bash
nix develop /home/chenyong/nix-config#go --command sh -c 'CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search'
```

## Answer Workflow

1. Prefer `search --format json` for AI-facing use because it returns `content`, `sources`, `sources_count`, and `model` in one object.
2. Use `fetch` when the user gives a specific URL or when a search result needs exact page content.
3. Use `map` before fetching many pages from a documentation site.
4. For independent questions, run independent CLI searches separately. Run follow-up searches only when the previous result creates a dependency.
5. Cite sources from the CLI JSON output. If sources are missing, say that the CLI answer did not provide sources.
6. Do not call another search backend for the same query unless the CLI fails, returns insufficient sources, or the user explicitly asks for comparison.

## Search Trigger Rules

Use the CLI instead of relying on model memory when:

- The user asks for latest/current/recent information.
- The answer depends on external facts, docs, APIs, laws, prices, schedules, releases, or public claims that may have changed.
- A specific library/framework/API usage question needs current official documentation.
- The user asks for evidence, citations, source comparison, or verification.
- Internal knowledge and external reality could plausibly differ.

If unsure whether information is current, search and state the uncertainty rather than guessing.

## Evidence Standards

- Treat CLI search answers as leads, not final authority.
- Prefer primary sources: official docs, standards, source repositories, API references, filings, papers, or original announcements.
- Key factual claims should have at least two independent sources when feasible. If only one credible source is available, say so.
- For conflicting sources, present both, compare credibility and dates, then state which is stronger or that the conflict is unresolved.
- For empirical or comparative conclusions, include confidence as `高` / `中` / `低` and briefly name the basis.
- Never fabricate citations. If a source lacks date, section, or author, cite only the fields that are actually available.

Citation style for final answers:

```text
[Organization/Author, date if available, title or section, URL]
```

## Expression Rules

- Be concise, direct, and information-dense.
- Use Markdown lists for discrete points and paragraphs for reasoning.
- Challenge flawed premises with evidence.
- State applicable conditions, scope boundaries, and known limitations for conclusions.
- Avoid greetings, filler, and emotional language.

## Configuration

Default config file:

```text
~/.config/grok-search/config.json
```

Use `config get` to inspect masked effective config. Use `config check` before diagnosing search failures.

The CLI supports `guda_api_key`/`guda_base_url` or explicit `grok_*`, `tavily_*`, and `firecrawl_*` settings. Environment variables override the config file.
