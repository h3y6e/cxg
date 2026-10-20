package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/h3y6e/cxg/internal/core/lint"
)

func TestRun(t *testing.T) {
	t.Parallel()

	unreadable := lint.New(func(string) ([]byte, error) {
		return nil, errors.New("permission denied")
	})

	t.Run("when args are nil, running the CLI prints root help and exits zero", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer

		options := Options{Version: "dev", Args: nil, Stdout: &stdout}

		// Act
		code := Run(options, unreadable)

		// Assert
		if code != 0 {
			t.Fatalf("Run() code = %d, want 0", code)
		}

		if !bytes.Contains(stdout.Bytes(), []byte("Usage:")) {
			t.Fatalf("stdout = %q, want root help", stdout.String())
		}
	})

	tests := []struct {
		name       string
		args       []string
		linter     linter
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name: "when lint receives no input, running the CLI reports a usage error and exits two",
			args: []string{"lint"},
			linter: linterFunc(func(lint.Input) (lint.Result, error) {
				return lint.Result{}, lint.ErrNoInput
			}),
			wantCode:   2,
			wantStderr: "lint: no input\n",
		},
		{
			name:       "when an unknown flag is provided, running the CLI reports it and exits two",
			args:       []string{"lint", "--unknown"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "unknown flag: --unknown\n",
		},
		{
			name:       "when an unknown command is provided, running the CLI reports it and exits two",
			args:       []string{"lnt"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "unknown command \"lnt\" for \"cxg\"\n",
		},
		{
			name:       "when lint finds violations, running the CLI reports them once and exits one",
			args:       []string{"lint", "--message", "bad message"},
			linter:     unreadable,
			wantCode:   1,
			wantStderr: "line 1 [invalid-subject] subject must match <type>(<scope>): <description>\n",
		},
		{
			name:       "when reading the file fails, running the CLI reports the cause and exits one",
			args:       []string{"lint", "COMMIT_EDITMSG"},
			linter:     unreadable,
			wantCode:   1,
			wantStderr: "reading file \"COMMIT_EDITMSG\": permission denied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			options := Options{
				Version: "dev",
				Args:    tt.args,
				Stdout:  &stdout,
				Stderr:  &stderr,
			}

			// Act
			code := Run(options, tt.linter)

			// Assert
			if code != tt.wantCode {
				t.Fatalf("Run() code = %d, want %d", code, tt.wantCode)
			}

			if stdout.String() != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.wantStdout)
			}

			if stderr.String() != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
