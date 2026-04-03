---
name: cxg
description: Use when an AI agent is about to create a git commit in a repo where cxg is available.
license: MIT
metadata:
  author: h3y6e
  version: "2026.4.0"
---

# cxg

Use this skill whenever an agent is about to create a new commit message. Route the message through `cxg` before it reaches git.

## Rule

- Never use `git commit -m ...` directly for a new message.
- Never write a handcrafted message straight to `git commit -F ...` when `cxg` is available.
- Use `cxg lint --fix ... | git commit -F -`.

If validation fails, fix the message and rerun `cxg`. Do not bypass it.

## Flow

```sh
cxg lint --fix \
  -m 'feat(auth): add login' \
  -m 'intent(auth): support enterprise SSO' \
  | git commit -F -
```

## Commit Format

```text
type(scope): subject line

action-type(scope): description
action-type(scope): description
```

Valid subject types:
`feat` `fix` `refactor` `perf` `test` `docs` `style` `build` `ci` `chore` `revert`

Valid action types:
`intent` `decision` `rejected` `constraint` `learned`

Subject rules:
- Follow Conventional Commits
- Maximum 72 characters
- No trailing period when using `--fix`

Body rules:
- Non-empty body lines must be valid action lines
- Free-form body text is not accepted
- `--fix` normalizes spacing and the subject/body gap

Rules:
- Use only action lines that carry signal
- Use the user's intent, not a restatement of the diff
- Include the reason in `rejected(...)`
- Do not invent context you do not have
- Trivial commits can stay subject-only

## Variants

Add a trailer:

```sh
cxg lint --fix \
  -m 'feat(auth): add login' \
  --trailer 'Co-authored-by: Alice <alice@example.com>' \
  | git commit -F -
```

Inspect machine-readable errors:

```sh
cxg lint --json -m 'bad message'
```

Lint a message file for a hook:

```sh
cxg lint .git/COMMIT_EDITMSG
```

## Before Committing

1. Check commit scope with `git diff --cached --stat`
2. Write the message with subject plus only the action lines you can support
3. Run `cxg lint --fix ... | git commit -F -`
4. If `cxg` rejects the message, fix it instead of bypassing validation
