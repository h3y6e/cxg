---
status: done
summary: Execution plan for `cxg commit` — a non-interactive command that auto-fixes, lints contextual commit messages, and then runs `git commit`
---

## Summary

- Add `cxg commit` as a non-interactive Cobra subcommand. Reuse the existing message resolution and lint flow, always apply the existing fix rules before linting, then invoke `git commit` only after linting succeeds. Keep `lint` behavior unchanged and support the global `--json` contract for structured commit results.

## Execution Context

- Runtime/Platform: Go 1.26, `github.com/h3y6e/cxg`
- Directory layout: `cmd/` owns Cobra wiring, stdout/stderr handling, and JSON marshaling; `internal/message/` owns input assembly and fix behavior; `internal/lint/` owns linting; `internal/git/` owns git subprocess helpers; `internal/commit/` owns commit orchestration; integration tests live in `cmd/`
- Dependencies: `github.com/spf13/cobra`, Go standard library `context`, `os/exec`, `bytes`, `encoding/json`
- Constraints: `commit` is non-interactive, supports only a defined subset of `git commit` flags, must not invoke git when linting fails, and must keep stdout as either normal git-facing output or a single JSON payload
- Validation: `go test ./...`, `mise run check`, `go build .`, `go build ./cmd/cxg`, and temp-repo integration tests for commit creation, failure propagation, and JSON output

## Key Decisions

- Reuse `message.Resolve`, `message.Fix`, and `lint.Validate` through an `internal/commit` orchestration layer rather than duplicating message preparation in `cmd/`
- Keep `cxg commit` opinionated: it always applies the existing fix rules before linting, while `cxg lint` keeps the explicit `--fix` toggle
- Add a small `internal/git` package modeled on `github.com/h3y6e/skills/internal/git`, using `exec.LookPath` + `exec.CommandContext` and focused helpers instead of open-coded git subprocesses throughout `cmd/`
- Keep `cmd/commit.go` thin: parse flags, call `internal/commit`, and format normal or JSON output; git execution and commit result shaping live under `internal/`
- Support only explicit non-interactive git flags from the spec: `--all`, `--amend`, `--allow-empty`, `--signoff`, `--author`, and `--date`
- Reject editor-based and lint-bypassing modes early instead of attempting full `git commit` parity
- In `--json` mode, capture git process output and encode commit results as JSON so stdout never mixes JSON and raw git output
- Update stale docs/spec references during polish so the repo no longer claims `cxg commit` is out of scope

## Format

- `- [ ] Txxx (P) [USn] description (path)`
- `(P)` marks a parallelizable task
- `[USn]` links the task to a user story in `specs/commit/spec.md`
- Tasks without `[USn]` are shared setup or infrastructure work

## Phase 1: Setup

- [x] T001 Add `commit` command scaffold and root/help coverage (`cmd/commit.go`, `cmd/root.go`, `cmd/root_test.go`)
- [x] T002 Define `commit` option structs and translation into an internal request type (`cmd/commit.go`, `internal/commit/service.go`)
- [x] T003 Add command-level tests for help output and unsupported-mode failures for git-side message-source, editor, and hook-skipping flags (`cmd/root_test.go`, `cmd/commit_test.go`)

### DoD

- [x] `cxg commit --help` prints usage and root help mentions `commit`
- [x] Unsupported interactive or lint-bypassing modes fail before any git subprocess starts
- [x] `go test ./cmd/...` passes

## Phase 2: Foundational

- [x] T004 Extract a shared internal message-preparation helper for resolve/fix/lint flow so `lint` and `commit` do not diverge (`internal/commit/prepare.go`, `cmd/lint.go`, `cmd/commit.go`)
- [x] T005 (P) Add `internal/git` helpers for locating git, checking repository state, and building commit subprocesses (`internal/git/git.go`, `internal/git/git_test.go`)
- [x] T006 Implement the `internal/commit` service and result model on top of `internal/git` (`internal/commit/service.go`, `internal/commit/service_test.go`)

### DoD

- [x] Validated message preparation is shared or aligned across `lint` and `commit`
- [x] Supported git flags map deterministically to subprocess arguments through tested `internal/git` and `internal/commit` helpers
- [x] `go test ./cmd/...` passes

## Phase 3: User Stories 1-2 (Priority: P1)

- [x] T007 [US1] Run `git commit` only after linting succeeds and pass the linted message as the commit message (`internal/commit/service.go`)
- [x] T008 (P) [US2] Apply default fixing before linting and commit the fixed message when it becomes lint-clean (`internal/commit/service.go`)
- [x] T009 [US1] [US2] Add unit and command tests for success, lint failure, file/stdin precedence, and fixed-message commits (`internal/commit/service_test.go`, `cmd/commit_test.go`)

### DoD

- [x] `cxg commit -m 'feat(auth): add login'` creates one commit in a temp repo with a staged change
- [x] Invalid input exits non-zero and creates no commit
- [x] `cxg commit` produces a normalized committed message by default
- [x] Success and lint-failure paths return the expected exit codes before P2 flag work starts
- [x] `go test ./cmd/...` passes

## Phase 4: User Stories 3-4 (Priority: P2)

- [x] T010 [US3] Wire `--all`, `--amend`, `--allow-empty`, `--signoff`, `--author`, and `--date` through the internal request and `git commit` execution path (`cmd/commit.go`, `internal/commit/service.go`)
- [x] T011 (P) [US4] Implement cmd-layer JSON responses from the internal result model with `valid`, `committed`, `message`, `errors`, and optional `gitError` while suppressing raw git stdout from stdout (`cmd/commit.go`, `cmd/commit_test.go`)
- [x] T012 [US3] [US4] Add integration tests for supported git flags, hook/repo-state failures, and JSON result shapes (`cmd/integration_test.go`, `cmd/commit_test.go`)

### DoD

- [x] Supported automation-oriented flags work in temp-repo tests
- [x] JSON output stays parseable with no mixed human-oriented git stdout
- [x] Git exit codes and failure details are preserved for callers
- [x] `go test ./...` passes

## Phase 5: Polish

- [x] T013 Update README and CLI-facing help text for `commit` usage and JSON behavior (`README.md`, `cmd/root.go`)
- [x] T014 (P) Align remaining docs with the new command and the clarified lint/commit split (`skills/cxg/SKILL.md`, related docs)

### DoD

- [x] Repo docs no longer claim `cxg commit` does not exist
- [x] `mise run check && go test ./... && go build . && go build ./cmd/cxg` passes
- [x] `specs/commit/plan.md` progress/results are synced before closing implementation
- [x] The implementation remains isolated enough that rollback means removing `commit` command wiring and docs, without touching lint rules

## Progress Log

<!--
- Task: Txxx
  Change: ...
  Doc Impact: local-only | plan-impacting | spec-impacting
  Validation: PASS/FAIL
  Next: Txxx
-->
- Task: planning
  Change: Created the initial approved execution plan for `cxg commit` from the approved spec, covering command wiring, git execution, JSON mode, tests, and doc alignment.
  Doc Impact: plan-impacting
  Validation: PASS (spec review, existing plan comparison, manual plan self-review)
  Next: T001
- Task: plan-review
  Change: Corrected task dependencies, removed invalid parallel markers, clarified the shared-flow refactor task, and fixed the Phase 3 DoD so it no longer depends on `--allow-empty` before that flag is implemented.
  Doc Impact: plan-impacting
  Validation: PASS (review feedback verification against `specs/commit/plan.md`, `cmd/lint.go`, and `specs/cli/spec.md`)
  Next: T001
- Task: architecture-review
  Change: Reworked the plan so `cmd/commit.go` stays thin and all commit orchestration moves into `internal/commit`, while also syncing the existing approved `cli` spec so it no longer conflicts with the approved `commit` spec.
  Doc Impact: plan-impacting
  Validation: PASS (review feedback verification against repo architecture guidance and spec cross-check)
  Next: T001
- Task: T001-T003
  Change: Added the `commit` command scaffold to the root command, defined the initial option surface, and covered help output plus explicit rejection of git-side editor, message-source, and hook-skipping flags.
  Doc Impact: local-only
  Validation: PASS (`go test ./cmd/...`)
  Next: T004
- Task: T004-T009
  Change: Added shared internal message preparation, introduced `internal/git` and `internal/commit`, moved commit orchestration out of `cmd/`, and implemented P1 commit behavior with unit and integration coverage.
  Doc Impact: local-only
  Validation: PASS (`go test ./internal/...`, `go test ./cmd/...`)
  Next: T010
- Task: T010-T014
  Change: Added supported git flag passthrough, JSON commit results, hook and repo-state failure handling, and updated README, root help, skill docs, and repo docs to reflect the new `commit` flow.
  Doc Impact: local-only
  Validation: PASS (`mise run check`, `go test ./...`, `go build .`, `go build ./cmd/cxg`)
  Next: close implementation
- Task: follow-up-default-fix
  Change: Removed the `commit`-only `--fix` flag, made `cxg commit` always auto-fix before linting inside `internal/commit`, and updated tests and docs to match the breaking CLI change while leaving `cxg lint --fix` unchanged.
  Doc Impact: local-only
  Validation: PASS (`go test ./cmd/... ./internal/commit/...`, `mise run check`, `go test ./...`, `go build ./cmd/cxg`, `git diff --check`)
  Next: close implementation
- Task: follow-up-refactor
  Change: Narrowed the default-fix policy back to `cmd/commit.go` by setting `PrepareRequest.Fix = true` at the CLI boundary, so `internal/commit` stays a generic execution layer instead of mutating caller input.
  Doc Impact: local-only
  Validation: PASS (`go test ./cmd/... ./internal/commit/...`)
  Next: close implementation
