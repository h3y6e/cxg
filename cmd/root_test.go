package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand_HelpIncludesLintCommand(t *testing.T) {
	cmd := NewRootCmd("dev")
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
	if !bytes.Contains([]byte(output), []byte("commit")) {
		t.Fatalf("expected help to mention commit command, got %q", output)
	}
}

func TestLintCommand_HelpPrintsUsage(t *testing.T) {
	cmd := NewRootCmd("dev")
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

func TestCommitCommand_HelpPrintsUsage(t *testing.T) {
	cmd := NewRootCmd("dev")
	cmd.SetArgs([]string{"commit", "--help"})

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
		t.Fatal("expected commit help output")
	}

	if !bytes.Contains([]byte(output), []byte("cxg commit")) {
		t.Fatalf("expected commit help to mention command usage, got %q", output)
	}
}
