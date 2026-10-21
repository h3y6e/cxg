---
status: approved
summary: Specification for `cxg recall`, a non-interactive filter that extracts contextual action lines from `git log` output piped on stdin.
---

## Context and Goals

- Context: `cxg lint` is the write side of contextual commits: `cxg lint | git commit -F -`. Captured action lines are only useful if agents can read them back when resuming work or before choosing an approach.
- Goal: `git log ... | cxg recall [query]` extracts action lines from the piped log and prints them per commit. Like `lint`, `recall` never calls git: the agent picks the revision range with `git log` arguments, and cxg only parses and filters. Narrative synthesis is the `cxg-recall` skill's job, not the CLI's.

## User Scenarios and Testing

### User Story 1 - Brief the current branch (Priority: P1)
An agent resuming work on a feature branch wants the intents, decisions, rejections, constraints, and learnings recorded on it.

Why this priority: Resuming work is the most frequent use.
Independent Test: `git log main..HEAD | cxg recall` prints the action lines of the branch commits.
Acceptance Scenarios:
1. Given piped `git log` output, `cxg recall` prints every commit that has action lines, in input order, as its short hash and subject followed by its action lines, and exits 0.
2. Commits without action lines are omitted.
3. Given empty stdin (e.g. a branch with no commits ahead), stdout is empty and the exit code is 0.

### User Story 2 - Query a scope (Priority: P1)
An agent about to work in a scope wants all context recorded for it across history.

Why this priority: Prevents re-exploring rejected approaches.
Independent Test: `git log --all | cxg recall auth` prints only action lines whose scope starts with `auth`.
Acceptance Scenarios:
1. `cxg recall auth` keeps action lines whose scope starts with `auth` (`auth`, `auth-tokens`).
2. The scope is matched literally: `cxg recall a+b` matches `a+b` and not `ab`.

### User Story 3 - Query an action type in a scope (Priority: P1)
An agent wants only the rejected approaches for a scope before proposing one.

Why this priority: `rejected` and `constraint` are the highest-value signals before a decision.
Independent Test: `git log --all | cxg recall 'rejected(auth)'` prints only `rejected(auth…)` lines.
Acceptance Scenarios:
1. `cxg recall 'rejected(auth)'` keeps action lines with type `rejected` and a scope starting with `auth`.
2. `cxg recall 'unknown(auth)'` is a usage error listing the valid action types.

### User Story 4 - Machine-readable output (Priority: P2)
An agent wants structured output.

Acceptance Scenarios:
1. `cxg recall --json` writes `{"schema":1,"commits":[{"hash":"...","subject":"...","actions":[{"type":"intent","scope":"auth","text":"..."}]}]}`; with no matches, `commits` is `[]`.
2. On usage or runtime errors, the same record carries `error` with `code`, `message`, and, for usage errors, a runnable `fix`: `{"schema":1,"commits":[],"error":{...}}`.
3. `CXG_JSON=1` has the same effect as `--json`, for every command.

### User Story 5 - Discover usage (Priority: P2)
An agent runs `cxg recall` without piping anything to learn how to use it.

Acceptance Scenarios:
1. `cxg recall`, `cxg recall -h`, and `cxg recall --help` with stdin on a terminal print help to stdout and exit 0.
2. Help shows examples for the three query forms piped from `git log`.
3. `cxg lint` with stdin on a terminal and no `-m` or file stays a usage error with empty stdout: its stdout feeds `git commit -F -`, which would commit help text.
4. Stdin counts as a terminal only when it is one; empty non-terminal stdin such as `</dev/null` is input.

### Edge Cases

- Non-empty stdin with no `commit <hash>` line (e.g. `--oneline` or a custom `--format`): usage error with `fix` `git log | cxg recall`.
- A query with stdin on a terminal: usage error with `fix` `git log --all | cxg recall -- <query>`, shell-quoted.
- Stdout closed early (`| head`): exit quietly.
- Decorations (`commit <hash> (HEAD -> main)`), `Merge:` headers, `--stat`, and `-p` output are tolerated; only the indented message block is read.
- Action lines are recognized with the same grammar `lint` validates; other body lines and trailers are ignored.

## Requirements

### Functional Requirements

- FR-001: `cxg recall` reads `git log` output in its default (medium) format from stdin.
- FR-002: A commit starts at a `commit <hash>` line; its message is the 4-space-indented block after the header, ending at the next non-indented, non-empty line.
- FR-003: The optional positional query is either `<scope>` or `<action>(<scope>)`; anything else is a usage error.
- FR-004: Scope matching is a literal prefix match.
- FR-005: Text output per matching commit: `<short-hash> <subject>`, then each matching action line indented by two spaces; commits separated by a blank line.
- FR-006: No match is not an error: empty stdout (or `"commits":[]`), exit 0.
- FR-007: Exit codes: 0 on success and help, 2 on usage error, 1 on runtime failure.
- FR-008: `recall` executes no external commands; its effect is `read`.
- FR-009: Every error carries a stable `code` (`usage` or `failure`) and the cause; usage errors add a runnable `fix`. Text mode prints them to stderr.
- FR-010: Every `--json` record of every command starts with `"schema":1`, versioned apart from the CLI; removing, renaming, or retyping a field bumps it.

### Key Entities

- **Commit**: hash, subject, and action lines parsed from one log entry.
- **ActionLine**: type, scope, and text; the same entity `lint` validates.
- **Query**: optional action type plus scope prefix.

## Success Criteria

- SC-001: `git log main..HEAD | cxg recall` lists the branch's action lines per commit.
- SC-002: `git log --all | cxg recall 'rejected(auth)'` lists only `rejected(auth…)` lines.
- SC-003: The `cxg-recall` skill ships with each release alongside `cxg`.

## Scope

### In Scope

- `recall` command, positional query, `--json`
- `skills/cxg-recall/SKILL.md`: when to recall, which `git log` range to pipe, how to brief from the output; it defers to the installed `cxg recall --help`
- Release wiring: ship `skills/` in every release archive, declare `skill/cxg` and `skill/cxg-recall` as `archive:` resources, list both in `.tagpr`, `.agents/skills`, and README
- `CXG_*` env vars for the existing command too; invalid values fail only commands that use them, never help or version

### Out of Scope

- Calling git, resolving base branches, or reading the working tree
- Summarizing or grouping action lines beyond per-commit output
- Parsing `--oneline`, `--format`, or `--pretty` variants other than the default
