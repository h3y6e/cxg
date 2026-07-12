package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

func TestRootCommand_HelpIncludesLintCommand(t *testing.T) {
	cmd := newRootCmd("dev", lint.New(os.ReadFile))
	cmd.SetArgs([]string{"--help"})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdout.String() + stderr.String()
	if output == "" {
		t.Fatal("expected help output")
	}

	if !bytes.Contains([]byte(output), []byte("lint")) {
		t.Fatalf("expected help to mention lint command, got %q", output)
	}
}

func TestLintCommand_HelpPrintsUsage(t *testing.T) {
	cmd := newRootCmd("dev", lint.New(os.ReadFile))
	cmd.SetArgs([]string{"lint", "--help"})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := stdout.String() + stderr.String()
	if output == "" {
		t.Fatal("expected lint help output")
	}

	if !bytes.Contains([]byte(output), []byte("cxg lint")) {
		t.Fatalf("expected lint help to mention command usage, got %q", output)
	}
}
