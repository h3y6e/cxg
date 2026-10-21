---
name: cxg-recall
description: Recalls the intents, decisions, rejected approaches, and constraints recorded in contextual commits. Use when starting a session, resuming work, or switching branches, and before proposing an approach in a scope.
compatibility: Requires git and cxg.
license: MIT
metadata:
  author: h3y6e
  version: "2026.10.1"
---

# cxg-recall

`cxg recall` reads plain `git log` output on stdin and prints each commit's action lines. Trust the installed `cxg recall --help` over this file.

## Pick the range

The range decides what you learn, and `cxg recall` never picks it for you.

| Situation | Command |
| --- | --- |
| Feature branch | `git log <base>..HEAD \| cxg recall`, with `<base>` the branch it was created from |
| Default branch | `git log -20 \| cxg recall` |
| Before working in a scope | `git log --all \| cxg recall <scope>` |
| Before proposing an approach | `git log --all \| cxg recall 'rejected(<scope>)'`, then `'constraint(<scope>)'` |

## Brief

Report only what the action lines say, and keep every `rejected` line with its reason: those prevent re-exploring dead ends. When intent changed across commits, show both. Empty output means no recorded context, not an error.

## Before proposing

When your approach matches a `rejected` line or breaks a `constraint` line, show that line to the user before proposing it. The reason may no longer hold, but the user decides that.
