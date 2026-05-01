# Startup Reading Order

Read these files before broad source-code exploration, planning, or non-trivial edits:

1. `llmdoc/must/project-context.md`
2. `llmdoc/must/development-rules.md`

Then read task-relevant docs:

- Go CLI changes: `llmdoc/architecture/go-cli-flow.md` and `llmdoc/reference/go-cli-commands.md`
- Configuration or install changes: `llmdoc/reference/configuration.md`
- Grok/Tavily/Firecrawl behavior: `llmdoc/architecture/external-integrations.md`
- Local validation: `llmdoc/guides/local-development.md`

If docs conflict with source code, treat source code as current behavior and update `llmdoc/memory/doc-gaps.md`.
