# cxg

**C**onte**x**tual **g**it commit CLI for AI Agents. Lints [Contextual Commits](https://github.com/berserkdisruptors/contextual-commits) messages and can create commits.

## Install

```sh
go install github.com/h3y6e/cxg@latest
npx skills add h3y6e/cxg
```

## Usage

If you're a human, there's nothing to do. AI agents will automatically use `cxg` when creating commits.

See [SKILL.md](skills/cxg/SKILL.md) for the commit format and rules.

```sh
# lint and create a commit
cxg commit -m 'feat(auth): add login' -m 'intent(auth): support social login'

# machine-readable commit result
cxg commit --json -m 'feat(auth): add login'
```

## Hook

If a human wants local commit-message linting, `cxg lint` can be wired into `commit-msg` manually:

```sh
cat > .git/hooks/commit-msg <<'EOF'
#!/bin/sh
cxg lint "$1"
EOF
chmod +x .git/hooks/commit-msg
```

## Development

Requires [mise](https://mise.jdx.dev/).

```sh
mise run check   # gofmt + go vet + staticcheck
mise run test    # Run tests
mise run build   # Build binary
```
