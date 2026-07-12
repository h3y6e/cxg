package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

type lintRunnerFunc func(lint.Input) (lint.Result, error)

func (f lintRunnerFunc) Run(input lint.Input) (lint.Result, error) {
	return f(input)
}

func TestLintCommandPort(t *testing.T) {
	t.Parallel()

	// Arrange
	var captured lint.Input
	runner := lintRunnerFunc(func(input lint.Input) (lint.Result, error) {
		captured = input
		return lint.Result{Message: "feat(auth): add login"}, nil
	})
	command := newRootCmd("dev", runner)
	command.SetArgs([]string{
		"lint",
		"--fix",
		"--message", "feat(auth): add login",
		"--trailer", "Co-authored-by: Alice <alice@example.com>",
	})
	command.SetIn(strings.NewReader("feat(stdin): ignored"))
	var stdout bytes.Buffer
	command.SetOut(&stdout)

	// Act
	err := command.Execute()

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.String() != "feat(auth): add login" {
		t.Fatalf("stdout = %q, want lint result", stdout.String())
	}
	if len(captured.Messages) != 1 || captured.Messages[0] != "feat(auth): add login" {
		t.Fatalf("input messages = %#v, want message flag", captured.Messages)
	}
	if captured.Stdin == nil {
		t.Fatal("input stdin = nil, want command stdin")
	}
	if captured.FilePath != "" {
		t.Fatalf("input file path = %q, want empty", captured.FilePath)
	}
	if len(captured.Trailers) != 1 || captured.Trailers[0] != "Co-authored-by: Alice <alice@example.com>" {
		t.Fatalf("input trailers = %#v, want trailer flag", captured.Trailers)
	}
	if !captured.Fix {
		t.Fatal("input fix = false, want true")
	}
}

func TestLintCommandInputSelection(t *testing.T) {
	t.Parallel()

	t.Run("when a file and stdin are available, executing lint selects only the file", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var captured lint.Input
		runner := lintRunnerFunc(func(input lint.Input) (lint.Result, error) {
			captured = input
			return lint.Result{Message: "feat(file): from file"}, nil
		})
		command := newRootCmd("dev", runner)
		command.SetArgs([]string{"lint", "COMMIT_EDITMSG"})
		command.SetIn(strings.NewReader("feat(stdin): ignored"))
		command.SetOut(&bytes.Buffer{})

		// Act
		err := command.Execute()

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if captured.FilePath != "COMMIT_EDITMSG" {
			t.Fatalf("input file path = %q, want COMMIT_EDITMSG", captured.FilePath)
		}
		if captured.Stdin != nil {
			t.Fatalf("input stdin = %T, want nil", captured.Stdin)
		}
	})

	t.Run("when stdin is a character device, executing lint does not read it", func(t *testing.T) {
		t.Parallel()

		// Arrange
		stdin, err := os.Open(os.DevNull)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		defer stdin.Close()

		var captured lint.Input
		runner := lintRunnerFunc(func(input lint.Input) (lint.Result, error) {
			captured = input
			return lint.Result{Message: "feat(auth): add login"}, nil
		})
		command := newRootCmd("dev", runner)
		command.SetArgs([]string{"lint"})
		command.SetIn(stdin)
		command.SetOut(&bytes.Buffer{})

		// Act
		err = command.Execute()

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if captured.Stdin != nil {
			t.Fatalf("input stdin = %T, want nil", captured.Stdin)
		}
	})
}
