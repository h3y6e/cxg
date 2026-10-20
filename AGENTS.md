`cxg` is a non-interactive Go CLI; `cxg lint` output is piped into `git commit -F -`.
Run `mise run check` and `mise run test` before finishing.
Hexagonal: `internal/core` imports only the standard library and itself.
When CLI flags or output change, update `skills/cxg/SKILL.md` (it ships with each release) and run `mise run docs` to regenerate `docs/cli.md`.
