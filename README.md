# cxg

**C**onte**x**tual **g**it commit CLI for AI Agents. Lints [Contextual Commits](https://github.com/berserkdisruptors/contextual-commits) messages.

## Install

```sh
go install github.com/h3y6e/cxg@latest
gh skill install h3y6e/cxg
```

## Usage

If you're a human, there's nothing to do. AI agents will automatically use `cxg` when creating commits.

```sh
# validate a message
cxg lint -m 'feat(auth): add login' -m 'intent(auth): support social login'

# machine-readable result
cxg lint --json -m 'feat(auth): add login'
```

See [SKILL.md](skills/cxg/SKILL.md) for the commit format and rules.

## Development

Requires [mise](https://mise.jdx.dev/).

```sh
mise run check   # gofmt + go vet + staticcheck
mise run test    # Run tests
mise run build   # Build binary
```
