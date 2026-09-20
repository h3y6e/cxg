# cxg

**C**onte**x**tual **g**it commit CLI for AI Agents. Lints [Contextual Commits](https://github.com/berserkdisruptors/contextual-commits) messages.

## Install

With [mise](https://mise.jdx.dev):

```sh
mise use -g --tool-option 'identity_prefix=https://github.com/h3y6e/cxg/' \
  --tool-option 'issuer=https://token.actions.githubusercontent.com' \
  packslip:cxg.h3y6e.com
mise skills sync -g
```

```toml
[tools]
"packslip:cxg.h3y6e.com" = { version = "latest", identity_prefix = "https://github.com/h3y6e/cxg/", issuer = "https://token.actions.githubusercontent.com" }
```

Artifacts stay on GitHub Releases; discovery uses the signed list at
`https://cxg.h3y6e.com/.well-known/packslip.json`. Domain projects need the
signer pin above — the domain name alone does not imply one.

Or install each piece yourself:

```sh
go install github.com/h3y6e/cxg@latest
gh skill install h3y6e/cxg
```

The agent skill is [`skills/cxg`](skills/cxg/SKILL.md).
With [packslip](https://packslip.dev) it matches the installed CLI version; with `gh skill install` you install that directory yourself.

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

Requires [mise](https://mise.jdx.dev).

```sh
mise run check   # Formatting, modernization, static analysis, and vulnerability checks
mise run test    # Run tests
mise run build   # Build binary
```
