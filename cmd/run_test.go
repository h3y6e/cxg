package cmd

import (
	"bytes"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("when lint receives no input, running the CLI reports the error and exits one", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		options := Options{
			Version: "dev",
			Args:    []string{"lint", "--message", "feat(auth): add login"},
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		runner := lintRunnerFunc(func(lint.Input) (lint.Result, error) {
			return lint.Result{}, lint.ErrNoInput
		})

		// Act
		code := Run(options, runner)

		// Assert
		if code != 1 {
			t.Fatalf("Run() code = %d, want 1", code)
		}
		if stdout.String() != "" {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		if stderr.String() != "lint: no input\n" {
			t.Fatalf("stderr = %q, want no-input diagnostic", stderr.String())
		}
	})

	t.Run("when lint finds violations, running the CLI does not duplicate its diagnostic", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		options := Options{
			Version: "dev",
			Args:    []string{"lint", "--message", "bad message"},
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		service := lint.New(func(string) ([]byte, error) {
			t.Fatal("read file must not be called")
			return nil, nil
		})

		// Act
		code := Run(options, service)

		// Assert
		if code != 1 {
			t.Fatalf("Run() code = %d, want 1", code)
		}
		expected := "line 1 [invalid-subject] subject must match <type>(<scope>): <description>\n"
		if stderr.String() != expected {
			t.Fatalf("stderr = %q, want %q", stderr.String(), expected)
		}
	})
}
