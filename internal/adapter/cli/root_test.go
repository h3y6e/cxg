package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/core/lint"
)

func TestRootCommand(t *testing.T) {
	t.Parallel()

	t.Run("when help is requested, executing the root prints help listing lint to stdout", func(t *testing.T) {
		t.Parallel()

		// Act
		stdout, stderr, err := executeRoot(t, lint.New(os.ReadFile), nil, "--help")

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !strings.Contains(stdout, "lint") {
			t.Fatalf("stdout = %q, want help mentioning lint", stdout)
		}

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("when lint help is requested, executing the root prints lint usage to stdout", func(t *testing.T) {
		t.Parallel()

		// Act
		stdout, stderr, err := executeRoot(t, lint.New(os.ReadFile), nil, "lint", "--help")

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !strings.Contains(stdout, "cxg lint") {
			t.Fatalf("stdout = %q, want lint usage", stdout)
		}

		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("when the version is requested, executing the root prints it to stdout", func(t *testing.T) {
		t.Parallel()

		// Arrange
		root := newRootCmd(&rootOptions{}, "v1.2.3", lint.New(os.ReadFile))
		root.SetArgs([]string{"--version"})

		var stdout strings.Builder
		root.SetOut(&stdout)

		// Act
		err := root.Execute()

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if stdout.String() != "cxg version v1.2.3\n" {
			t.Fatalf("stdout = %q, want version", stdout.String())
		}
	})

	t.Run("when the usage spec is requested, executing the root prints a spec declaring lint", func(t *testing.T) {
		t.Parallel()

		// Act
		stdout, _, err := executeRoot(t, lint.New(os.ReadFile), nil, "--usage-spec")

		// Assert
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !strings.Contains(stdout, "bin cxg\n") || !strings.Contains(stdout, "cmd lint ") {
			t.Fatalf("stdout = %q, want usage spec with lint", stdout)
		}
	})
}

func TestTerminalStdin(t *testing.T) {
	t.Parallel()

	t.Run("when lint gets no message on a terminal, executing it fails as a usage error and prints nothing for git commit", func(t *testing.T) {
		t.Parallel()

		// Act
		stdout, _, err := executeOnTerminal(t, "lint")

		// Assert
		if exitStatus(err) != 2 {
			t.Fatalf("exit status = %d (error %v), want 2", exitStatus(err), err)
		}

		if stdout != "" {
			t.Fatalf("stdout = %q, want empty", stdout)
		}
	})
}
