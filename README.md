# cxg

`cxg` is a non-interactive CLI for validating and formatting contextual commit messages.

It never calls git. On success it writes the validated message to stdout so an agent can pipe it to `git commit -F -`. On failure it exits with code `1`.

## Install

```sh
go install github.com/h3y6e/cxg@latest
```

For local development:

```sh
go build -o cxg .
```

## Usage

Validate a message passed with `-m`:

```sh
cxg lint -m 'feat(auth): add login'
```

Build a message with multiple `-m` flags:

```sh
cxg lint \
  -m 'feat(auth): add login' \
  -m 'intent(auth): support social login'
```

Append trailers:

```sh
cxg lint \
  -m 'feat(auth): add login' \
  -m 'intent(auth): support social login' \
  --trailer 'Co-authored-by: Alice <alice@example.com>'
```

Read from a file path, which is useful for `commit-msg` hooks:

```sh
cxg lint .git/COMMIT_EDITMSG
```

Read from stdin:

```sh
printf '%s\n' 'feat(auth): add login' | cxg lint
```

Emit machine-readable JSON:

```sh
cxg lint --json -m 'feat(auth): add login'
```

Auto-fix common formatting issues before validating:

```sh
cxg lint --fix -m $'feat(auth): add login.\nintent(auth): support social login'
```

## Git Pipelines

Pipe the validated message into git:

```sh
cxg lint -m 'feat(auth): add login' | git commit --allow-empty -F -
```

Use `--fix` in the same pipeline:

```sh
cxg lint --fix -m $'feat(auth): add login.\nintent(auth): support social login' | git commit --allow-empty -F -
```

## Commit-Message Hook

Create `.git/hooks/commit-msg`:

```sh
#!/bin/sh
cxg lint "$1"
```

Then make it executable:

```sh
chmod +x .git/hooks/commit-msg
```

## Development

```sh
mise run check
go test ./...
```
