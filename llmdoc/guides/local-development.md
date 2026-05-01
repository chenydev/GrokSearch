# Local Development

## Setup

Use any Go toolchain compatible with the `go.mod` version. The repository does not commit `.envrc`; if a contributor uses Nix or direnv, they should keep local environment files untracked.

## Build the CLI

Build a local static binary:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search
```

`dist/` is ignored and should not be committed.

## Basic Validation

Run:

```bash
go test ./...
go vet ./...
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/grok-search ./cmd/grok-search
```

## Investigation Tips

- Start with `internal/cli/root.go` to understand command registration and runtime flow.
- Check `internal/config/config.go` before editing README installation examples or environment variable docs.
- Check `internal/sources/sources.go` before changing citation/source formatting behavior.
