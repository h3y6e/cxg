# cxg

Commit message linter for AI agents. Validates [Contextual Commits](https://github.com/berserkdisruptors/contextual-commits) format.

## Install

```sh
go install github.com/h3y6e/cxg@v2026.4.0
npx skills add h3y6e/cxg
```

## Usage

If you're a human, there's nothing to do. AI agents will automatically use `cxg` when creating commits.

See [SKILL.md](.agents/skills/cxg/SKILL.md) for the commit format and rules.

## Development

Requires [mise](https://mise.jdx.dev/).

```sh
mise run check   # gofmt + go vet + staticcheck
mise run test    # Run tests
mise run build   # Build binary
```
