package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

func TestRootCommand_HelpIncludesLintCommand(t *testing.T) {
	// Arrange
	cmd := newRootCmd("dev", lint.New(os.ReadFile))
	cmd.SetArgs([]string{"--help"})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if stdout.String() == "" {
		t.Fatal("expected help output")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("lint")) {
		t.Fatalf("expected help to mention lint command, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestLintCommand_HelpPrintsUsage(t *testing.T) {
	// Arrange
	cmd := newRootCmd("dev", lint.New(os.ReadFile))
	cmd.SetArgs([]string{"lint", "--help"})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if stdout.String() == "" {
		t.Fatal("expected lint help output")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("cxg lint")) {
		t.Fatalf("expected lint help to mention command usage, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootCommandVersion(t *testing.T) {
	// Arrange
	cmd := newRootCmd("v1.2.3", lint.New(os.ReadFile))
	cmd.SetArgs([]string{"--version"})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.String() != "cxg version v1.2.3\n" {
		t.Fatalf("stdout = %q, want version", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
