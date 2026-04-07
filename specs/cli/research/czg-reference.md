# czg Reference

Source: https://cz-git.qbb.sh/cli/

## What czg Is

czg is a standalone interactive CLI for composing Conventional Commit messages. It replaces `commitizen` (148 packages, 102MB) with a zero-dependency binary (~1.32MB) that prompts the user through type → scope → subject → body → footer interactively and then calls `git commit`.

## Key Commands

```sh
czg                     # Interactive commit wizard
czg ai                  # Generate subject via OpenAI
czg break               # Append ! for breaking changes
czg emoji               # Output with emoji
czg --retry             # Re-submit last message
czg --alias <key>       # Submit predefined message
czg --config <path>     # Use custom config file
```

## Commit Format (Output)

czg outputs standard Conventional Commits:

```
type(scope): subject

body

footer
```

Default type list: `feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert`

## Configuration (`.czrc` / `cz.config.js`)

Relevant fields for cxg's lint logic:

| Field | Default | Notes |
|-------|---------|-------|
| `types` | 11 built-in | Each has `value`, `name`, optional `emoji` |
| `scopes` | `[]` | Predefined scope list |
| `allowCustomScopes` | `true` | |
| `allowEmptyScopes` | `true` | |
| `upperCaseSubject` | `false` | |
| `minSubjectLength` | `0` | |
| `maxSubjectLength` | `100` | czg default; cxg uses 72 per spec |
| `useEmoji` | `false` | |
| `breaklineNumber` | `100` | body line-wrap threshold |

## Where cxg Diverges from czg

| Dimension | czg | cxg |
|-----------|-----|-----|
| Interaction model | Interactive prompt wizard | Non-interactive; reads `-m`, stdin, or file |
| Git integration | Calls `git commit` directly | Never touches git; pipes to stdout |
| Primary user | Human developer at terminal | AI agent in a pipeline |
| Message composition | Wizard builds message | Message pre-assembled by caller |
| Scope | Configurable per project | Fixed type list (spec FR-005) |
| Body format | Free-form | Action lines linted (`intent`, `decision`, etc.) |
| Fix mode | N/A | `--fix` auto-normalizes before linting |
| Output | Git commit created | Message to stdout; errors to stderr |

## What cxg Takes from czg

- Conventional Commits type list as the canonical baseline (feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert)
- The idea of a lightweight, zero-setup CLI that fits into any workflow
- `--retry`-like pattern influence: cxg's `-m` flag mirrors `git commit -m` so agents use it the same way they would `git commit`

## Why cxg Is Not czg

czg is a human-facing wizard that *creates* commits. cxg is an agent-facing validator that *checks and formats* messages before an agent calls `git commit -F -`. The two are complementary: czg for humans, cxg for agents.
