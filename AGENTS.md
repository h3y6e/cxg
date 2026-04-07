# AGENTS.md

`cxg` is a non-interactive Go CLI for contextual commit workflows.

- `cxg lint` lints and formats messages and writes the linted message to stdout for lint-only workflows.
- `cxg commit` fixes and lints the message first, then invokes `git commit` directly.

## Build / Lint / Test

Task runner: mise (see `mise.toml`). All CI runs through mise tasks.

```sh
mise run build          # go build -o cxg .
mise run check          # test -z "$(gofmt -l .)" && go vet ./... && staticcheck ./...
mise run fmt            # gofmt -w .
mise run test           # go test ./...
```

## Architecture Conventions

- `cmd/` handles CLI concerns (flags, I/O, JSON marshaling). 
- Delegates all logic to `internal/`.
- No global state. Commands accept options structs.
