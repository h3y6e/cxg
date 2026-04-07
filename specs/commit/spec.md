---
status: approved
summary: Requirements and scenarios for a non-interactive `cxg commit` command that lints contextual commit messages before invoking `git commit`.
---

## Context and Goals

- Context: Today `cxg` lints and formats commit messages, but the caller must still pipe the result into `git commit -F -`. The approved `cli` spec explicitly keeps `cxg commit` out of scope.
- Goal: Add `cxg commit` as a non-interactive command that preserves `cxg`'s message composition and linting behavior, applies the existing fix rules by default, and creates the git commit only when linting succeeds.

## User Scenarios and Testing

### User Story 1 - Lint and create a commit in one command (Priority: P1)
An AI agent wants one command that lints a contextual commit message and creates the git commit without manual piping.

Why this priority: This is the main reason to add `cxg commit` instead of requiring `cxg lint ... | git commit -F -`.
Independent Test: In a temporary git repo with a staged file change, `cxg commit -m "feat(auth): add login"` creates exactly one commit with that message.
Acceptance Scenarios:
1. In a repository with a staged change. When `cxg commit -m "feat(auth): add login"` is run with a lint-clean message, `git commit` is invoked and the commit is created with the linted message.
2. In a repository with a staged change. When `cxg commit -m "bad message"` is run, `cxg` prints lint errors to stderr, exits non-zero, and does not create a commit.
3. In a repository where the caller provides multiple `-m` flags and `--trailer`, when `cxg commit` succeeds, the stored commit message matches the same composition rules as `cxg lint`.

### User Story 2 - Auto-fix before committing by default (Priority: P1)
An AI agent wants `cxg commit` to normalize a rough generated message automatically and create the commit only after the normalized message passes linting.

Why this priority: Agent-generated messages are a common source of formatting mistakes that should not require a separate preflight step.
Independent Test: In a temporary git repo with a staged file change, `cxg commit -m "feat(auth): add login.\nintent(auth): support social login"` creates a commit with a normalized blank line and no trailing period in the subject.
Acceptance Scenarios:
1. When a message becomes valid after default fixing, `cxg commit` uses the fixed message as the committed message.
2. When a message still fails linting after default fixing, `cxg commit` prints lint errors, exits non-zero, and does not invoke `git commit`.
3. `cxg commit` does not expose a separate `--fix` flag; fixing is always part of the commit flow.

### User Story 3 - Preserve common non-interactive git commit workflows (Priority: P2)
An AI agent wants to keep using common `git commit` behaviors from automation without falling back to manual piping.

Why this priority: `cxg commit` must fit existing automated workflows without becoming a full replacement for all git porcelain.
Independent Test: In a repo with a tracked file modification, `cxg commit --all -m "docs(readme): refresh usage"` stages the tracked change via git and creates the commit.
Acceptance Scenarios:
1. When the caller uses supported non-interactive commit options such as `--all`, `--amend`, `--allow-empty`, `--signoff`, `--author`, or `--date`, `cxg commit` applies them to the resulting git commit.
2. When git rejects the commit because of repo state or hooks, `cxg commit` surfaces git's failure and does not mask it with a generic success message.
3. When the caller requests an interactive or message-source-conflicting mode, `cxg commit` fails early with a clear error instead of opening an editor or bypassing linting.

### User Story 4 - Receive machine-readable commit results (Priority: P2)
An AI agent wants to use the global `--json` flag with `cxg commit` and receive structured output about linting and commit execution.

Why this priority: `--json` is already a global CLI contract, and automation should not need a separate parsing mode for `commit`.
Independent Test: In a temporary git repo with a staged file change, `cxg commit --json -m "feat(auth): add login"` creates the commit and returns a JSON object with `valid`, `committed`, `message`, and `errors`.
Acceptance Scenarios:
1. When `cxg commit --json -m "feat(auth): add login"` succeeds, stdout contains only JSON with `valid`, `committed`, `message`, and `errors`.
2. When linting fails under `cxg commit --json`, stdout contains only JSON describing the lint errors, stderr remains empty unless JSON encoding itself fails, and no commit is created.
3. When git fails after linting under `cxg commit --json`, stdout contains only JSON with `valid`, `committed`, `message`, `errors`, and `gitError`.

### Edge Cases

- Lint failure must stop before any `git commit` process starts.
- If multiple message sources are present, precedence is repeated `-m`, then stdin, then a file path argument.
- If git reports `nothing to commit`, `cxg commit` must surface that git failure to the caller, either through the normal git output path or the JSON result shape.
- Existing commit hooks still run unless changed by an explicitly supported git option.
- `cxg commit` remains non-interactive in the initial scope.
- When `--json` is set, raw `git commit` stdout must not be mixed into stdout alongside the JSON payload.
- Git-side message-source or lint-bypassing flags are rejected explicitly instead of being passed through.

## Requirements

### Functional Requirements

- FR-001: `cxg commit` accepts message input from the same sources and precedence order as `cxg lint`: repeated `-m`, then stdin, then a file path argument.
- FR-002: `cxg commit` accepts repeated `--trailer` values and composes multiple `-m` values plus trailers using the same rules as `cxg lint`.
- FR-003: `cxg commit` lints the composed message with the same contextual-commit rules as `cxg lint`.
- FR-004: `cxg commit` always applies the existing fix rules before linting and does not expose a separate `--fix` flag.
- FR-005: When linting fails, `cxg commit` writes lint errors to stderr, exits non-zero, and does not invoke `git commit`.
- FR-006: When linting succeeds, `cxg commit` invokes `git commit` and uses the linted message as the commit message.
- FR-007: `cxg commit` supports common non-interactive `git commit` options needed in automation: `--all`, `--amend`, `--allow-empty`, `--signoff`, `--author`, and `--date`.
- FR-008: If the git commit process fails after linting, `cxg commit` preserves git's exit code and surfaces git's failure details to the caller.
- FR-009: `cxg commit` rejects interactive or lint-bypassing git modes, including `--edit`/`-e`, `--file`/`-F`, `--reuse-message`/`-C`, `--reedit-message`/`-c`, `--template`/`-t`, and `--no-verify`.
- FR-010: When the global `--json` flag is set, `cxg commit` returns a machine-readable JSON object with `valid`, `committed`, `message`, and `errors` for success and lint failure.
- FR-011: In `--json` mode, git execution failures are returned via an additional `gitError` object, and raw `git commit` stdout is suppressed from stdout.
- FR-012: The initial `CommitResult` JSON shape does not include commit SHA or other derived commit metadata.
- FR-013: `cxg commit` does not add new staging UX; file selection remains the responsibility of existing git behavior.

### Key Entities

- **CommitInput**: user-provided message parts, repeated `--trailer` values, and supported git commit options before linting.
- **LintedCommitMessage**: the fully composed commit message that passed `cxg` linting and is safe to hand to git.
- **GitCommitRequest**: the supported non-interactive git commit options to apply together with the linted message.
- **CommitResult**: machine-readable output with `valid`, `committed`, `message`, and `errors`, plus optional `gitError` when git fails after linting.
- **GitError**: structured git failure details containing the subprocess exit code and captured stderr/stdout needed by automation.

## Success Criteria

- SC-001: `cxg commit -m "feat(auth): add login"` creates a commit from staged changes without requiring manual piping.
- SC-002: An invalid message never creates a git commit.
- SC-003: `cxg commit` can normalize a rough message and create the resulting commit in one step without an extra flag.
- SC-004: Common automation-oriented git commit options continue to work through `cxg commit`.
- SC-005: Hook failures and repo-state failures are visible to the caller without losing the original git error.
- SC-006: `cxg commit --json` returns parseable JSON with `valid` and `committed` fields and without interleaved human-oriented git output.

## Scope

### In Scope

- A new non-interactive `commit` subcommand
- Reuse of existing message input, repeated `--trailer`, composition, linting, and default fix behavior
- Common non-interactive `git commit` options needed for automation
- Global `--json` support for structured commit results
- Clear propagation of lint failures and git failures

### Out of Scope

- Interactive editors, prompts, or wizards
- Staging arbitrary paths or adding new `git add` behavior
- Full `git commit` flag parity
- Commit SHA or other derived commit metadata in JSON output
- Hook installation helpers
- AI-based commit message generation

## Acceptance Criteria

- `cxg commit` can create a commit from staged changes using the same message composition rules as `cxg lint`.
- Messages that fail linting never start `git commit`, even when staged changes are present.
- `cxg commit` normalizes a rough message and commits the normalized result by default.
- Supported automation-oriented git flags behave as documented, while unsupported interactive or lint-bypassing flags fail clearly.
- `cxg commit --json` returns a stable machine-readable result shape without mixing raw git stdout into stdout.
