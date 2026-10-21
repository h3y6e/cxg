package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/core/lint"
)

type linterFunc func(lint.Input) (lint.Result, error)

func (f linterFunc) Lint(input lint.Input) (lint.Result, error) {
	return f(input)
}

func TestLintCommandInput(t *testing.T) {
	t.Parallel()

	pipe := strings.NewReader("feat(stdin): ignored")

	tests := []struct {
		name  string
		args  []string
		stdin io.Reader
		want  lint.Input
	}{
		{
			name: "when flags are given, executing lint passes them and piped stdin to the linter",
			args: []string{
				"lint",
				"--fix",
				"--message", "feat(auth): add login",
				"--trailer", "Co-authored-by: Alice <alice@example.com>",
			},
			stdin: pipe,
			want: lint.Input{
				Messages: []string{"feat(auth): add login"},
				Stdin:    pipe,
				Trailers: []string{"Co-authored-by: Alice <alice@example.com>"},
				Fix:      true,
			},
		},
		{
			name:  "when a file and stdin are available, executing lint selects only the file",
			args:  []string{"lint", "COMMIT_EDITMSG"},
			stdin: strings.NewReader("feat(stdin): ignored"),
			want:  lint.Input{FilePath: "COMMIT_EDITMSG"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var got lint.Input

			linter := linterFunc(func(input lint.Input) (lint.Result, error) {
				got = input

				return lint.Result{}, nil
			})

			// Act
			_, _, err := executeRoot(t, linter, tt.stdin, tt.args...)

			// Assert
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("input = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLintCommandOutput(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte("feat(auth): add login\n\nintent(auth): support social login\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		stdin      io.Reader
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{
			name: "when the message is valid, executing lint writes it to stdout",
			args: []string{
				"lint",
				"-m", "feat(auth): add login",
				"-m", "intent(auth): support social login",
				"-m", "decision(auth): keep OAuth optional",
				"--trailer", "Co-authored-by: Alice <alice@example.com>",
			},
			wantStdout: "feat(auth): add login\n\nintent(auth): support social login\ndecision(auth): keep OAuth optional\n\nCo-authored-by: Alice <alice@example.com>",
		},
		{
			name:       "when the message is invalid, executing lint writes violations to stderr and fails",
			args:       []string{"lint", "-m", "bad message"},
			wantStatus: 1,
			wantStderr: "line 1 [invalid-subject] subject must match <type>(<scope>): <description> with type one of feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert\n",
		},
		{
			name:       "when a file is given, executing lint writes its content without the final newline",
			args:       []string{"lint", file},
			wantStdout: "feat(auth): add login\n\nintent(auth): support social login",
		},
		{
			name:       "when message flags and stdin are given, executing lint prefers the flags",
			args:       []string{"lint", "-m", "feat(auth): add login"},
			stdin:      strings.NewReader("fix(auth): from stdin"),
			wantStdout: "feat(auth): add login",
		},
		{
			name:       "when json is requested for a valid message, executing lint writes a valid record",
			args:       []string{"lint", "--json", "-m", "feat(auth): add login"},
			wantStdout: `{"schema":1,"valid":true,"message":"feat(auth): add login","errors":[]}` + "\n",
		},
		{
			name:       "when json is requested for an invalid message, executing lint writes its violations and fails",
			args:       []string{"lint", "--json", "-m", "bad message"},
			wantStatus: 1,
			wantStdout: `{"schema":1,"valid":false,"errors":[{"line":1,"code":"invalid-subject","message":"subject must match \u003ctype\u003e(\u003cscope\u003e): \u003cdescription\u003e with type one of feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert"}]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			stdout, stderr, err := executeRoot(t, lint.New(os.ReadFile), tt.stdin, tt.args...)

			// Assert
			if status := exitStatus(err); status != tt.wantStatus {
				t.Fatalf("exit status = %d (error %v), want %d", status, err, tt.wantStatus)
			}

			if stdout != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout, tt.wantStdout)
			}

			if stderr != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr, tt.wantStderr)
			}
		})
	}
}

func executeRoot(t *testing.T, linter linter, stdin io.Reader, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	return execute(t, &rootOptions{}, linter, stdin, args...)
}

func executeOnTerminal(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	return execute(t, &rootOptions{terminal: true}, lint.New(os.ReadFile), nil, args...)
}

func execute(t *testing.T, opts *rootOptions, linter linter, stdin io.Reader, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	root := newRootCmd(opts, "dev", linter)
	root.SetArgs(args)

	if stdin != nil {
		root.SetIn(stdin)
	}

	var out, errOut bytes.Buffer

	root.SetOut(&out)
	root.SetErr(&errOut)

	err = root.Execute()

	return out.String(), errOut.String(), err
}

func exitStatus(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[exitError](err); ok {
		return exitErr.status
	}

	return -1
}
