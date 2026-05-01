# llmdoc Index

This directory contains stable, reusable project knowledge for future coding sessions. Temporary investigation notes stay in `.llmdoc-tmp/` and are not part of the stable documentation set.

## Startup

- `startup.md`: required reading order before broad project work.

## Must Read

- `must/project-context.md`: project identity, runtime shape, and core ownership boundaries.
- `must/development-rules.md`: repo-specific development rules and known sharp edges.

## Overview

- `overview/project-overview.md`: product purpose, package structure, and current implementation scope.

## Architecture

- `architecture/go-cli-flow.md`: Go CLI command layout and runtime flow.
- `architecture/external-integrations.md`: Grok, Tavily, Firecrawl, config, retry, and logging boundaries.

## Guides

- `guides/local-development.md`: local setup, inspection, and basic validation workflow.

## Reference

- `reference/configuration.md`: environment variables, model persistence, and config precedence.
- `reference/go-cli-commands.md`: Go CLI command inventory and output conventions.

## Memory

- `memory/reflections/`: process reflections owned by the llmdoc reflection workflow.
- `memory/reflections/2026-05-02-go-cli-v1.md`: reflection from the first Go CLI migration pass.

## Routing Hints

- Start with `startup.md` for every non-trivial task.
- Use `reference/configuration.md` before changing environment handling, setup docs, logging, or model switching.
- Use `architecture/go-cli-flow.md` and `reference/go-cli-commands.md` before changing the Go CLI.
