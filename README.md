# cxg

**C**onte**x**tual **g**it commit CLI for AI Agents. Lints [Contextual Commits](https://github.com/berserkdisruptors/contextual-commits) messages and recalls their action lines.

## Install

With [mise](https://mise.jdx.dev):

```sh
mise use -g packslip:github.com/h3y6e/cxg
mise skills sync -g
```

Or install each piece yourself:

```sh
go install github.com/h3y6e/cxg/cmd/cxg@latest
gh skill install h3y6e/cxg
```

The agent skills are [`skills/cxg`](skills/cxg/SKILL.md) for commits and [`skills/cxg-recall`](skills/cxg-recall/SKILL.md) for recalling context.

## Usage

If you're a human, there's nothing to do. AI agents will automatically use `cxg` when creating commits.

```sh
# commit only a valid message
cxg lint -m 'feat(auth): add login' -m 'intent(auth): support social login' | git commit -F -

# recall action lines from git log
git log main..HEAD | cxg recall
git log --all | cxg recall 'rejected(auth)'
```

See [SKILL.md](skills/cxg/SKILL.md) for the commit format and rules, and [docs/usage.md](docs/usage.md) for every command and flag.

## Development

Requires [mise](https://mise.jdx.dev).

```sh
mise run check   # golangci-lint, go fix, and govulncheck
mise run test    # Run tests
mise run build   # Build binary
```
