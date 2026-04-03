---
status: approved
summary: Specification for cxg, a non-interactive CLI that builds, validates, and formats contextual commit messages and outputs them to stdout for AI agents.
---

## Context and Goals

- Context: Contextual commits extend Conventional Commits by adding action lines (intent, decision, rejected, constraint, learned) to the commit body to preserve "why the code was written." cxg is a CLI tool that validates and formats these messages.
- Goal: `cxg` never touches git. It validates and formats messages, outputs them to stdout, and lets agents pipe the result to git. It can also be used as a `commit-msg` hook. Unlike czg (an interactive git wrapper), cxg is a non-interactive message processor. The main command is `cxg lint`; auto-fix is available via the `--fix` flag.

## User Scenarios and Testing

### User Story 1 - Validate a commit message and pipe it to git (Priority: P1)
An AI agent wants to validate a message and, if valid, stream it to stdout and pass it to `git commit -F -`.

Why this priority: This is the core use case of cxg.
Independent Test: `cxg lint -m "feat(auth): add login" | git commit -F -` succeeds.
Acceptance Scenarios:
1. Running `cxg lint -m "..."` with a valid message outputs the message body to stdout and exits with code 0.
2. Running `cxg lint -m "..."` with an invalid message outputs an error to stderr, leaves stdout empty, and exits with code 1.
3. Running `cxg lint --json -m "..."` with a valid message outputs `{"valid":true,"message":"..."}` to stdout.

### User Story 2 - Use as a commit-msg hook (Priority: P1)
Register cxg as a `commit-msg` hook to automatically validate messages on `git commit`.

Why this priority: Hooks integrate naturally into git workflows.
Independent Test: Write `cxg lint "$1"` in `.git/hooks/commit-msg`; a commit with an invalid message is aborted.
Acceptance Scenarios:
1. `cxg lint <filepath>` reads the message from the file and validates it. Exits 0 if valid; outputs an error to stderr and exits 1 if invalid.
2. When git runs it as a `commit-msg` hook, a commit with an invalid message is aborted.

### User Story 3 - Build a message with multiple -m flags and --trailer (Priority: P1)
Assemble a message using multiple `-m` flags and `--trailer`, the same way as `git commit`.

Why this priority: A git-compatible CLI lets agents use it intuitively.
Independent Test: `cxg lint -m "feat(auth): add login" -m "intent(auth): social login" | cat` outputs the combined message.
Acceptance Scenarios:
1. Multiple `-m` flags produce a message joined by blank lines (same behavior as `git commit`).
2. `--trailer "Co-authored-by: Alice <alice@example.com>"` is appended at the end.
3. If the combined message passes validation, it is written to stdout.

### User Story 4 - Auto-fix a message with --fix before validating (Priority: P2)
An AI agent wants to auto-fix a rough generated message, validate it, and pass it to git.

Why this priority: Agent-generated messages often have missing blank lines or trailing spaces.
Independent Test: `cxg lint --fix -m "feat(auth): add login\nintent(auth): ..."` outputs a blank-line-normalized, validated message.
Acceptance Scenarios:
1. If the gap between the subject and body is 0 or 2+ lines, normalize it to exactly one blank line, validate, and output to stdout if valid.
2. Blank lines within the body are removed.
3. Trailing whitespace on lines is removed.
4. A trailing period on the subject is removed (`feat(auth): add login.` → `feat(auth): add login`).
5. Leading whitespace on body lines is removed (normalizes indented action lines).
6. If the message is still invalid after fixing, output an error to stderr and exit 1 (stdout is empty).
4. The pipeline `cxg lint --fix -m "..." | git commit -F -` works end-to-end.

### User Story 5 - Switch output format (Priority: P2)
Receive structured output via `--json`. Default is stdout passthrough.

Why this priority: Agents need to parse error details from JSON.
Acceptance Scenarios:
1. `cxg lint --json` returns `{"valid":true,"message":"..."}` or `{"valid":false,"errors":[...]}`.
2. Without `--json`, only the message body is written to stdout when valid.

### Edge Cases

- A subject line only (no body) is valid and is output to stdout as-is.
- If both `-m` and stdin are provided, `-m` takes precedence.
- Multiple `-m` joining: one blank line between each `-m`.
- `--trailer` is appended at the end of the message, preceded by a blank line.
- `cxg lint --fix` auto-fixes first, then validates; only valid output is written to stdout.

## Requirements

### Functional Requirements

- FR-001: `cxg lint` accepts a message from `-m` (multiple allowed), stdin, or a file path argument, and validates it.
- FR-002: When valid, write the message body to stdout (without `--json`).
- FR-003: When invalid, write errors to stderr and leave stdout empty.
- FR-004: The `--json` flag returns `{"valid":bool,"message":"...","errors":[...]}` to stdout.
- FR-005: The subject line is validated against Conventional Commits format. Fixed type list: feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert.
- FR-006: Action lines are validated as `action-type(scope): description`. Valid types: intent, decision, rejected, constraint, learned.
- FR-007: Exit code is 0 on success and 1 on validation error.
- FR-008: Multiple `-m` flags are joined by blank lines (same as `git commit`).
- FR-009: `--trailer <token>:<value>` is appended at the end of the message.
- FR-010: The `--fix` flag auto-fixes the message before validating. If still invalid after fixing, output to stderr and exit 1. Fix operations: (1) normalize the subject-body gap to exactly one blank line (add if missing, collapse if more than one), (2) remove blank lines within the body, (3) remove trailing whitespace, (4) remove a trailing period from the subject, (5) remove leading whitespace from body lines.
- FR-013: A subject line longer than 72 characters is an error. Body line length is not validated.
- FR-011: `cxg lint <filepath>` reads and validates a file (for commit-msg hook use).
- FR-012: `cxg commit` / `cxg validate` commands are not provided.

### Key Entities

- **CommitMessage**: struct composed of subject, body lines, and trailers.
- **ActionLine**: struct with action-type, scope, and description.
- **ValidationError**: struct with line number, error code, and message.

## Success Criteria

- SC-001: `cxg lint -m "..." | git commit -F -` succeeds.
- SC-002: Writing `cxg lint "$1"` in `.git/hooks/commit-msg` makes the hook work.
- SC-003: All examples in SKILL.md are judged valid.
- SC-004: The `cxg lint --fix -m "..." | git commit -F -` pipeline works.

## Scope

### In Scope

- `lint` command (validation + stdout passthrough)
- `lint --fix` flag (auto-fix + validation)
- Multiple `-m` flags, `--trailer` flag
- File path argument (for hook use)
- `--json` global flag
- Go implementation using cobra

### Out of Scope

- Calling git commands (cxg never touches git)
- Automatic hook installation (`cxg hook install`, etc.)
- AI-based auto-generation
- `recall` / `diff` commands
- MCP server

## Acceptance Criteria

- [ ] `go build ./cmd/cxg/` succeeds
- [ ] `go test ./...` passes
- [ ] `cxg lint -m "feat(auth): add login" | git commit -F -` creates a commit
- [ ] `cxg lint /path/to/COMMIT_EDITMSG` works as a hook
- [ ] `cxg lint --fix -m "feat: add\nintent(x): ..."` normalizes blank lines and outputs the result
- [ ] Multiple `-m` flags are joined by blank lines
