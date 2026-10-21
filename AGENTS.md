`cxg` is a non-interactive Go CLI that never runs git: `cxg lint` output is piped into `git commit -F -`, and `git log` output is piped into `cxg recall`.
Run `mise run check` and `mise run test` before finishing.
Hexagonal: `internal/core` imports only the standard library and itself.
When CLI flags or output change, update the affected `skills/*/SKILL.md`; they ship with each release.
