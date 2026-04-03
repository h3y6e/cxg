# Project Structure & Tooling Reference

Source: `~/ghq/github.com/h3y6e/skills`

## Directory Layout

```
.
├── main.go                    # Entry point; var version = "dev"
├── cmd/
│   ├── root.go                # NewRootCmd(version), Execute(version) error
│   ├── <command>.go           # Flag declarations + runXxx() dispatch
│   └── <command>_test.go
├── internal/
│   └── <package>/
│       ├── <file>.go
│       └── <file>_test.go
├── mise.toml
├── .goreleaser.yaml
├── .tagpr
└── .gitignore
```

## main.go Pattern

```go
var version = "dev"

func main() {
    if err := cmd.Execute(version); err != nil {
        os.Exit(1)
    }
}
```

## cmd/root.go Pattern

```go
func NewRootCmd(version string) *cobra.Command {
    root := &cobra.Command{
        Use:           "cxg",
        SilenceUsage:  true,
        SilenceErrors: true,
        Version:       resolveVersion(version),
    }
    root.AddCommand(newLintCmd())
    return root
}

func Execute(version string) error {
    return NewRootCmd(version).Execute()
}
```

`resolveVersion` reads `debug.ReadBuildInfo()` and appends the short VCS revision.

## mise.toml

```toml
[tools]
"aqua:dominikh/go-tools/staticcheck" = "2026.1"
go = "1.26.1"

[tasks.build]
run = "go build -o cxg ."

[tasks.check]
run = "gofmt -l . && go vet ./... && staticcheck ./..."

[tasks.fmt]
run = "gofmt -w ."

[tasks.test]
run = "go test ./..."

[tasks.prerelease]
run = ["go mod tidy", "git add go.mod go.sum"]
```

## .goreleaser.yaml

```yaml
version: 2
builds:
  - env: [CGO_ENABLED=0]
    flags: [-trimpath]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
archives:
  - format_overrides:
      - goos: windows
        formats: [zip]
changelog:
  use: github-native
release:
  use_existing_draft: true
```

## .tagpr

```toml
[tagpr]
releaseBranch = main
versionFile = -
vPrefix = true
release = draft
changelog = false
calendarVersioning = YYYY.MM.MICRO
command = "mise run prerelease"
```

## .gitignore

Standard Go gitignore: binary name, `*.test`, `coverage.*`, `go.work`, `.env`.

## Notes

- `cmd/root.go` declares all subcommands inline (flag wiring); each `cmd/<name>.go` holds the `runXxx()` logic.
- `SilenceUsage` + `SilenceErrors` on root: error printing is the caller's responsibility.
- Tests live alongside source (`_test.go` in same package or `_test` suffix package).
- No `vendor/` directory; standard `go.sum`.
