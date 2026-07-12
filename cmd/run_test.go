package cmd

import (
	"bytes"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("when lint receives no input, running the CLI reports a usage error and exits two", func(t *testing.T) {
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
		if code != 2 {
			t.Fatalf("Run() code = %d, want 2", code)
		}
		if stdout.String() != "" {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		if stderr.String() != "lint: no input\n" {
			t.Fatalf("stderr = %q, want no-input diagnostic", stderr.String())
		}
	})

	t.Run("when an unknown flag is provided, running the CLI reports it and exits two", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		options := Options{
			Version: "dev",
			Args:    []string{"lint", "--unknown"},
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
		if code != 2 {
			t.Fatalf("Run() code = %d, want 2", code)
		}
		if stdout.String() != "" {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		if stderr.String() != "unknown flag: --unknown\n" {
			t.Fatalf("stderr = %q, want unknown-flag diagnostic", stderr.String())
		}
	})

	t.Run("when an unknown command is provided, running the CLI reports it and exits two", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		options := Options{
			Version: "dev",
			Args:    []string{"unknown"},
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
		if code != 2 {
			t.Fatalf("Run() code = %d, want 2", code)
		}
		if stdout.String() != "" {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		expected := "unknown command \"unknown\" for \"cxg\"\n"
		if stderr.String() != expected {
			t.Fatalf("stderr = %q, want %q", stderr.String(), expected)
		}
	})

	t.Run("when args are nil, running the CLI treats them as empty", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		options := Options{
			Version: "dev",
			Args:    nil,
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
		if code != 0 {
			t.Fatalf("Run() code = %d, want 0", code)
		}
		if !bytes.Contains(stdout.Bytes(), []byte("Usage:")) {
			t.Fatalf("stdout = %q, want root help", stdout.String())
		}
		if stderr.String() != "" {
			t.Fatalf("stderr = %q, want empty", stderr.String())
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
