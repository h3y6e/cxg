package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
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

	t.Run("when the json record cannot be written, running the CLI reports the failure on stderr and exits one", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stderr bytes.Buffer

		options := Options{
			Version: "dev",
			Args:    []string{"lint", "--json", "-m", "feat(auth): add login"},
			Stdout:  failingWriter{},
			Stderr:  &stderr,
		}

		// Act
		code := Run(options, lint.New(os.ReadFile))

		// Assert
		if code != 1 {
			t.Fatalf("Run() code = %d, want 1", code)
		}

		if stderr.String() != "no space left on device\n" {
			t.Fatalf("stderr = %q, want write failure", stderr.String())
		}
	})
}

func TestRunReportsFailures(t *testing.T) {
	t.Parallel()

	unreadable := lint.New(func(string) ([]byte, error) {
		return nil, errors.New("permission denied")
	})

	tests := []runCase{

		{
			name:       "when an unknown flag is provided, running the CLI points to the command help and exits two",
			args:       []string{"lint", "--unknown"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "unknown flag: --unknown\nFix: cxg lint --help\n",
		},
		{
			name:       "when a misspelled command is provided, running the CLI suggests the closest one and exits two",
			args:       []string{"lnt"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "unknown command \"lnt\" for \"cxg\"\n\nDid you mean this?\n\tlint\nFix: cxg --help\n",
		},
		{
			name:       "when lint finds violations, running the CLI reports them once and exits one",
			args:       []string{"lint", "--message", "bad message"},
			linter:     unreadable,
			wantCode:   1,
			wantStderr: "line 1 [invalid-subject] subject must match <type>(<scope>): <description> with type one of feat, fix, refactor, perf, test, docs, style, build, ci, chore, revert\n",
		},
		{
			name:       "when reading the file fails, running the CLI reports the cause and exits one",
			args:       []string{"lint", "COMMIT_EDITMSG"},
			linter:     unreadable,
			wantCode:   1,
			wantStderr: "reading file \"COMMIT_EDITMSG\": permission denied\n",
		},
		{
			name:       "when json is requested and an argument error occurs, running the CLI writes an error record and exits two",
			args:       []string{"lint", "--json", "a", "b"},
			linter:     unreadable,
			wantCode:   2,
			wantStdout: `{"schema":1,"valid":false,"errors":[],"error":{"code":"usage","message":"accepts at most 1 arg(s), received 2","fix":"cxg lint --help"}}` + "\n",
		},
		{
			name:       "when flag parsing fails, running the CLI reports the error as text even with --json and exits two",
			args:       []string{"lint", "--json", "--fix=invalid"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "invalid argument \"invalid\" for \"--fix\" flag: strconv.ParseBool: parsing \"invalid\": invalid syntax\nFix: cxg lint --help\n",
		},
		{
			name:       "when json is requested and reading the file fails, running the CLI writes an error record and exits one",
			args:       []string{"lint", "--json", "COMMIT_EDITMSG"},
			linter:     unreadable,
			wantCode:   1,
			wantStdout: `{"schema":1,"valid":false,"errors":[],"error":{"code":"failure","message":"reading file \"COMMIT_EDITMSG\": permission denied"}}` + "\n",
		},
	}

	runCases(t, tests)
}

func TestRunReadsEnv(t *testing.T) {
	t.Parallel()

	unreadable := lint.New(func(string) ([]byte, error) {
		return nil, errors.New("permission denied")
	})

	tests := []runCase{
		{
			name:       "when CXG_JSON is set, running the CLI writes json without the flag",
			args:       []string{"lint", "-m", "feat(auth): add login"},
			env:        []string{"CXG_JSON=1"},
			linter:     unreadable,
			wantStdout: `{"schema":1,"valid":true,"message":"feat(auth): add login","errors":[]}` + "\n",
		}, {
			name:       "when an env var belongs to another command's flag, running the CLI ignores it",
			args:       []string{"--version"},
			env:        []string{"CXG_FIX=maybe"},
			linter:     unreadable,
			wantStdout: "cxg version dev\n",
		}, {
			name:       "when an env var is invalid and the version is requested, running the CLI still prints the version",
			args:       []string{"--version"},
			env:        []string{"CXG_JSON=maybe"},
			linter:     unreadable,
			wantStdout: "cxg version dev\n",
		}, {
			name:       "when one env var is invalid and CXG_JSON is set, running the CLI reports the error as a json record",
			args:       []string{"lint", "-m", "feat(auth): add login"},
			env:        []string{"CXG_FIX=maybe", "CXG_JSON=1"},
			linter:     unreadable,
			wantCode:   2,
			wantStdout: `{"schema":1,"valid":false,"errors":[],"error":{"code":"usage","message":"invalid CXG_FIX: strconv.ParseBool: parsing \"maybe\": invalid syntax","fix":"unset CXG_FIX"}}` + "\n",
		}, {
			name:       "when CXG_JSON is not a boolean, running the CLI reports a usage error and exits two",
			args:       []string{"lint", "-m", "feat(auth): add login"},
			env:        []string{"CXG_JSON=maybe"},
			linter:     unreadable,
			wantCode:   2,
			wantStderr: "invalid CXG_JSON: strconv.ParseBool: parsing \"maybe\": invalid syntax\nFix: unset CXG_JSON\n",
		}}

	runCases(t, tests)
}

func TestRunPrintsHelpDespiteInvalidEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		terminal bool
	}{
		{name: "when help for recall is requested, running the CLI prints it", args: []string{"help", "recall"}},
		{name: "when recall gets no query on a terminal, running the CLI prints its help", args: []string{"recall"}, terminal: true},
	}

	for _, tt := range tests {
		t.Run(tt.name+" even with an invalid env var", func(t *testing.T) {
			t.Parallel()

			// Arrange
			var stdout bytes.Buffer

			options := Options{
				Version:       "dev",
				Args:          tt.args,
				Env:           []string{"CXG_JSON=maybe"},
				Stdout:        &stdout,
				StdinTerminal: tt.terminal,
			}

			// Act
			code := Run(options, lint.New(os.ReadFile))

			// Assert
			if code != 0 {
				t.Fatalf("Run() code = %d, want 0", code)
			}

			if !strings.Contains(stdout.String(), "Usage:\n  cxg recall") {
				t.Fatalf("stdout = %q, want recall help", stdout.String())
			}
		})
	}
}

type runCase struct {
	name       string
	args       []string
	env        []string
	linter     linter
	wantCode   int
	wantStdout string
	wantStderr string
}

func runCases(t *testing.T, tests []runCase) {
	t.Helper()

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
				Env:     tt.env,
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

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("no space left on device")
}
