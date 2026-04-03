# AGENTS.md

`cxg` is a non-interactive Go CLI that validates and formats contextual commit
messages. It never calls git itself -- on success it writes the validated message
to stdout so the caller can pipe it to `git commit -F -`.

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
