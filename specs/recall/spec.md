---
status: approved
summary: Requirements and scenarios for the `cxg recall` subcommand
---

## Context and Goals

- Context: `cxg` captures contextual action lines (`intent`, `decision`, `rejected`, `constraint`, `learned`) in commit bodies. Once captured, developers need a way to query and review that context — especially when resuming work, switching branches, or investigating past decisions in a specific scope.
- Goal: Provide a single CLI subcommand (`cxg recall`) that extracts and presents contextual action lines from git history in three modes: full branch briefing, scope query, and action+scope query.

## User Scenarios and Testing

### User Story 1 — Branch Briefing (Priority: P1)

A developer starts a session on a feature branch and wants a summary of what's been decided, rejected, and constrained so far.

Why this priority: The most common use case — resuming work.
Independent Test: Run `cxg recall` on a branch with contextual commits and verify the output contains extracted action lines grouped by type.
Acceptance Scenarios:
1. On a feature branch with contextual commits ahead of the base branch. When running `cxg recall` with no arguments, action lines from those commits are extracted and printed to stdout, grouped by type (intent, decision, rejected, constraint, learned).
2. On a feature branch with zero commits ahead of the base branch. When running `cxg recall`, the output indicates no contextual history on this branch.
3. On the default branch. When running `cxg recall`, recent commits (up to 20) are scanned for action lines.
4. On any branch where commits exist but contain no action lines. When running `cxg recall`, the output states that no contextual action lines were found.

### User Story 2 — Scope Query (Priority: P1)

A developer wants to review all contextual decisions ever made for a particular scope (e.g. `auth`).

Why this priority: Key to avoiding re-exploration of rejected approaches.
Independent Test: Run `cxg recall auth` in a repo with action lines scoped to `auth` and verify all matching lines appear.
Acceptance Scenarios:
1. A repo has contextual commits with scopes `auth`, `auth-tokens`, `auth-library`. When running `cxg recall auth`, all action lines whose scope starts with `auth` are returned (prefix matching), grouped by action type.
2. No commits match the given scope. When running `cxg recall payments`, the output states no matches were found.

### User Story 3 — Action+Scope Query (Priority: P1)

A developer wants to see only rejected approaches for a scope before proposing a new one.

Why this priority: Prevents wasted exploration on known dead ends.
Independent Test: Run `cxg recall rejected(auth)` and verify only `rejected(auth…)` lines appear.
Acceptance Scenarios:
1. A repo has `rejected(auth)` and `decision(auth)` lines. When running `cxg recall rejected(auth)`, only `rejected(auth…)` lines are returned, each with the originating commit subject for provenance.
2. No commits match the action+scope combination. When running `cxg recall constraint(payments)`, the output states no matches were found.

### Edge Cases

- Argument looks like `action(scope)` but action is not one of the five recognized types → parsed as action+scope; the query returns no matches (no fallback to scope mode).
- Scope contains special regex characters → must be handled safely (literal matching).
- Repository has no commits at all → exit gracefully with an informative message.
- Merge commits or commits with empty bodies → skip silently, no errors.

## Requirements

### Functional Requirements

- FR-001: Argument parsing distinguishes three modes: no-arg (default), bare word (scope), and `action(scope)` pattern.
- FR-002: The five recognized action types are `intent`, `decision`, `rejected`, `constraint`, `learned`.
- FR-003: Default mode determines the base branch for comparison. Try the upstream tracking branch first, then the nearest local branch by commit distance, then fall back to the repository default branch.
- FR-004: Default mode on a feature branch extracts action lines from commits between the base branch and HEAD.
- FR-005: Default mode on the default branch extracts action lines from the most recent 20 commits.
- FR-006: Scope query searches the full repository history (`--all`) for action lines matching the given scope prefix.
- FR-007: Action+scope query searches the full repository history for lines matching the exact action type and scope prefix.
- FR-008: Output is written to stdout. Errors to stderr.
- FR-009: All output is plain text, suitable for terminal display. The root-level `--json` flag (shared across all subcommands) provides structured JSON output as an alternative.

### Key Entities

- **Action line**: A line in a commit body matching `^(intent|decision|rejected|constraint|learned)(scope): text`. The five action types carry distinct semantics.
- **Scope**: The parenthesized token in an action line. Queries use prefix matching (e.g. `auth` matches `auth`, `auth-tokens`).
- **Base branch**: The closest ancestor branch used to determine which commits belong to the current feature branch.

## Success Criteria

- SC-001: `cxg recall` with no arguments outputs action lines from the current branch context.
- SC-002: `cxg recall <scope>` outputs all matching action lines across the repository, grouped by action type.
- SC-003: `cxg recall <action>(<scope>)` outputs only matching lines with commit provenance.
- SC-004: All three modes exit 0 on success, including when no matches are found (no-match is not an error).

## Scope

### In Scope

- The three invocation modes (default, scope, action+scope).
- Action line extraction and grouping.
- Base branch detection logic.
- Plain text output to stdout (plus inherited `--json` from the root command).

### Out of Scope

- Synthesized narrative or conversational summaries (output is structured data, not prose).
- Unstaged/staged diff analysis (the CLI tracks working tree state as boolean metadata but does not analyze diffs).
- Proactive checking (that is a skill/agent concern, not a CLI concern).

## Acceptance Criteria

- [ ] `cxg recall` runs in all three modes and exits 0.
- [ ] Action lines are correctly extracted by regex from commit bodies.
- [ ] Scope query uses prefix matching across full history.
- [ ] Action+scope query filters by both action type and scope prefix.
- [ ] Base branch detection follows the priority: upstream → nearest branch → default.
- [ ] No output corruption on repos with no contextual commits (graceful fallback).
- [ ] `mise run test` passes.
- [ ] `mise run check` passes.
