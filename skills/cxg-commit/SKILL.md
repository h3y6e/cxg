---
name: cxg-commit
description: Writes contextual commit messages and lints them with cxg before committing. Use when creating a git commit.
compatibility: Requires git and cxg.
license: MIT
metadata:
  author: h3y6e
  version: "2026.10.2"
---

# cxg-commit

Commit through `cxg lint`, which prints the message only when it is valid; add `--fix` to normalize formatting first. Trust the installed `cxg lint --help` over this file.

```sh
cxg lint -m 'feat(auth): add social login' -m 'intent(auth): users asked to sign in with Google' | git commit -F -
```

## Action lines

The body holds only action lines, `<action-type>(<scope>): <description>`. They are the record `cxg recall` reads back later, so write what the diff cannot show:

- `intent` — what the user wanted and why, in their terms
- `decision` — the approach chosen over alternatives, and why
- `rejected` — an approach considered and dropped, always with the reason
- `constraint` — a hard limit or dependency that shaped the change
- `learned` — an API quirk or undocumented behavior worth knowing next time

Write only what this session's conversation or investigation supports; when nothing does, commit the subject alone. A restated diff or a guessed reason misleads whoever recalls it.

Scopes are concept labels such as `auth` or `session-store`. Reuse the names earlier commits use, since `cxg recall <scope>` matches by prefix.

## Examples

```text
fix(button): correct alignment on mobile viewport
```

```text
feat(notifications): add weekly email digest

intent(notifications): users want one weekly summary instead of per-event emails
rejected(digest-schedule): daily digest — users said it was still too noisy
constraint(email-provider): SendGrid batch API caps a call at 1000 recipients
```
