---
name: cxg-recall
description: Use when resuming work, switching branches, or before proposing approaches in a scope area.
compatibility: Requires git and cxg.
license: MIT
metadata:
  author: h3y6e
  version: "2026.4.3"
---

# cxg recall

Reconstruct development context from contextual commit history. Run before starting work to understand what has been decided, rejected, and learned.

## When to Use

- **Session start** — run `cxg recall` to get a branch context briefing.
- **Before proposing an approach** — run `cxg recall <scope>` to check for rejections and constraints.
- **Checking specific signals** — run `cxg recall "rejected(<scope>)"` to see what was already tried and discarded.

## Modes

### 1. Branch context (no arguments)

```sh
cxg recall
```

Outputs action lines grouped by type (rejected/constraint first), each prefixed with its scope:

```
rejected:
  - auth: auth0-sdk — session model incompatible with redis store
constraint:
  - auth: Redis session TTL 24h, tokens must refresh within window
intent:
  - auth: Add Google as first social login provider
decision:
  - auth: passport.js for multi-provider flexibility
learned:
  - auth: passport-google needs explicit offline_access scope for refresh tokens
```

### 2. Scope query

```sh
cxg recall auth
```

Searches all branches for action lines whose scope starts with the given term. Prefix matching: `auth` matches `auth`, `auth-tokens`, `auth-library`.

### 3. Action+scope query

```sh
cxg recall "rejected(auth)"
```

Returns only the specified action type for a scope, with source commit provenance.

## JSON output

```sh
cxg recall --json
cxg recall --json auth
cxg recall --json "rejected(auth)"
```

## Interpreting Output

Action types, in priority order:

- **rejected** — what was discarded and why (**highest-value signal** — never re-propose without acknowledging)
- **constraint** — hard limits that shape implementation
- **intent** — what the user wanted and why
- **decision** — which approach was chosen
- **learned** — API quirks and undocumented behavior

**Before proposing a new approach**, always check `rejected` and `constraint` for that scope. If a prior rejection exists for your proposed approach, surface it to the user before proceeding.
