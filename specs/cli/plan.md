---
status: done
summary: Execution plan for cxg CLI — Go-based contextual commit message validator with lint, fix, and JSON output
---

## Summary

Build `cxg lint` as a non-interactive Go CLI using cobra. Accepts messages from `-m`, stdin, or file path; validates Conventional Commits subject + action lines; writes valid output to stdout (or JSON); writes errors to stderr. Supports `--fix` auto-correction and `--trailer` appending.

## Execution Context

- Runtime/Platform: Go 1.26, `github.com/h3y6e/cxg`
- Directory layout: `main.go` at root, `cmd/cxg/main.go` as a spec-compatible alternate entrypoint, `cmd/` for cobra commands, `internal/message/` and `internal/lint/` for logic
- Dependencies: `github.com/spf13/cobra`
- Tooling: `mise.toml` (go 1.26, staticcheck); tasks: build, check (gofmt + go vet + staticcheck), fmt, test, prerelease
- Release: `.goreleaser.yaml` (CGO_ENABLED=0, -trimpath, -ldflags version); `.tagpr` (calendarVersioning)
- Constraints: cxg never calls git. stdout is clean (only message or JSON). stderr is errors only.
- Validation: `mise run check && go test ./...`; manual pipeline `cxg lint -m "..." | git commit -F -`

## Key Decisions

- `main.go` at root with `var version = "dev"`; version injected via ldflags `-X main.version={{.Version}}`
- `cmd.Execute(version) error` pattern; `NewRootCmd(version)` with `SilenceUsage: true, SilenceErrors: true`
- Single command `lint`; no `commit` or `validate` subcommands (FR-012)
- Fixed Conventional Commits type list: feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert (FR-005)
- Fixed action line types: intent, decision, rejected, constraint, learned (FR-006)
- Multiple `-m` joined by blank lines matching `git commit` behavior (FR-008)
- `--fix` mutates input before validation; invalid-after-fix still exits 1 (FR-010)

## Phase 1: Setup

- [x] T001 Init Go module (`go mod init github.com/h3y6e/cxg`), create `main.go` with `cmd.Execute(version)`
- [x] T002 Create `mise.toml` (go 1.26, staticcheck; tasks: build, check, fmt, test, prerelease)
- [x] T003 Create `.goreleaser.yaml` (CGO_ENABLED=0, -trimpath, `-X main.version={{.Version}}`)
- [x] T004 Create `.tagpr` (calendarVersioning, release=draft) and `.gitignore`
- [x] T005 Scaffold cobra root command (`cmd/root.go`) and `lint` subcommand stub (`cmd/lint.go`)

### DoD

- [x] `go build .` produces `./cxg` binary
- [x] `./cxg lint --help` prints usage
- [x] `mise run check` passes (gofmt clean, go vet, staticcheck)

## Phase 2: Foundational

- [x] T006 Define `CommitMessage`, `ActionLine`, `ValidationError` structs (`internal/message/types.go`)
- [x] T007 (P) Implement input resolution: `-m` flags → stdin → file path; `-m` takes precedence (FR-001, FR-011) [US1, US2]
- [x] T008 (P) Implement message assembly: join multiple `-m` with blank lines, append `--trailer` (FR-008, FR-009) [US3]
- [x] T009 Implement message parser: split assembled string into subject, body lines, trailers (`internal/message/parser.go`)

### DoD

- [x] Unit tests: multi-`-m` join, trailer append, input precedence
- [x] `go test ./internal/...` passes

## Phase 3: Lint Command — P1 Stories

- [x] T010 Implement subject validation: type list, optional scope, description, 72-char limit (FR-005, FR-013) [US1]
- [x] T011 (P) Implement action line validation: `action-type(scope): description` format (FR-006) [US1]
- [x] T012 Wire lint command: valid → write message to stdout + exit 0; invalid → write errors to stderr + exit 1 (FR-002, FR-003, FR-007) [US1, US2]

### DoD

- [x] `cxg lint -m 'feat(auth): add login'` exits 0 and prints message [US1]
- [x] `cxg lint -m 'bad message'` exits 1 and stderr shows error [US1]
- [x] `cxg lint <filepath>` reads file and validates (commit-msg hook) [US2]
- [x] `cxg lint -m 'feat: x' -m 'intent(x): why'` outputs joined message [US3]
- [x] `go test ./...` passes

## Phase 4: P2 Features

- [x] T013 Implement `--fix` transformations: normalize subject-body gap, remove intra-body blanks, strip trailing whitespace, remove trailing subject period, strip leading whitespace from body lines (FR-010) [US4]
- [x] T014 (P) Implement `--json` output: `{"valid":bool,"message":"...","errors":[...]}` to stdout (FR-004) [US5]

### DoD

- [x] `cxg lint --fix -m 'feat: add.\nintent(x): ...'` normalizes and outputs fixed message [US4]
- [x] Invalid-after-fix exits 1 with stderr, empty stdout [US4]
- [x] `cxg lint --json -m '...'` outputs valid JSON [US5]
- [x] `go test ./...` passes

## Phase 5: Polish

- [x] T015 Add integration tests covering full pipelines from acceptance criteria (SC-001 to SC-004)
- [x] T016 (P) Write README with install, usage examples, commit-msg hook setup

### DoD

- [x] `cxg lint -m 'feat(auth): add login' | git commit -F -` creates a commit (SC-001)
- [x] `cxg lint "$1"` works as `commit-msg` hook (SC-002)
- [x] All examples from contextual-commit SKILL.md validate as valid (SC-003)
- [x] `cxg lint --fix -m '...' | git commit -F -` pipeline works end-to-end (SC-004)
- [x] `mise run check && go test ./...` passes with no failures

## Progress Log

<!--
- Task: Txxx
  Change: ...
  Doc Impact: local-only | plan-impacting | spec-impacting
  Validation: PASS/FAIL
  Next: Txxx
-->
- Task: T001-T005
  Change: Initialized the Go module, release/tooling config, and Cobra CLI scaffold with `lint` help coverage.
  Doc Impact: local-only
  Validation: PASS (`go test ./cmd/...`, `go build .`, `./cxg lint --help`, `mise run check`)
  Next: T006
- Task: T006-T009
  Change: Added the core message model plus input resolution, multi-`-m`/trailer assembly, and subject/body/trailer parsing with focused tests.
  Doc Impact: local-only
  Validation: PASS (`go test ./internal/message/...`, `go test ./internal/...`)
  Next: T010
- Task: T010-T012
  Change: Implemented subject and action-line validation plus `lint` command stdout/stderr/exit behavior for flag, stdin-precedence, and file-path inputs.
  Doc Impact: local-only
  Validation: PASS (`go test ./internal/lint/...`, `go test ./cmd/...`)
  Next: T013
- Task: T013-T014
  Change: Added message auto-fix normalization and root-level JSON output while preserving stdout/stderr separation and exit semantics.
  Doc Impact: local-only
  Validation: PASS (`go test ./internal/message/...`, `go test ./cmd/...`, `go test ./...`)
  Next: T015
- Task: T015-T016
  Change: Added end-to-end git pipeline and hook integration tests, validated contextual-commit skill examples, documented install/usage/hook workflows in README, and added `cmd/cxg/main.go` so `go build ./cmd/cxg` also succeeds.
  Doc Impact: local-only
  Validation: PASS (`go test ./cmd/...`, `mise run check`, `go test ./...`, `go build .`, `go build ./cmd/cxg`)
  Next: close implementation
