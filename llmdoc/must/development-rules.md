# Development Rules

## Default Workflow

- Read `llmdoc/startup.md` and the MUST docs before broad code search.
- Keep temporary investigation notes under `.llmdoc-tmp/investigations/`.
- Update stable docs when source behavior, public tool contracts, configuration, or workflow knowledge changes.
- Preserve user changes in the working tree; inspect before editing files that may already be modified.

## Source of Truth

- Treat `internal/cli/root.go` as the source of truth for the Go CLI command surface.
- Treat `internal/config/config.go` as the source of truth for Go CLI config precedence and supported environment variables.
- Treat README files as product/user docs that may lag source behavior.

## Known Sharp Edges

- Go CLI `search` stores only the latest result in user cache for `last`; it does not implement MCP-style long-lived source sessions.
- The CLI uses Cobra but avoids Viper; config behavior is intentionally implemented in `internal/config`.
- `dist/` is a local build output and must stay ignored.
- `.envrc` is a local environment file and must stay ignored; do not commit machine-specific Nix setup.

## Go CLI Validation

Use the available Go toolchain. In this workspace, `direnv exec .` may work if a local `.envrc` exists; otherwise use the user's own Go environment.

```bash
nix develop /home/chenyong/nix-config#go --command go test ./...
nix develop /home/chenyong/nix-config#go --command go vet ./...
nix develop /home/chenyong/nix-config#go --command sh -c 'go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search'
```
