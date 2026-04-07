package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestCommitCommand_RejectsUnsupportedGitModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "edit", args: []string{"commit", "--edit", "-m", "feat(auth): add login"}},
		{name: "file", args: []string{"commit", "--file", "msg.txt", "-m", "feat(auth): add login"}},
		{name: "reuse", args: []string{"commit", "--reuse-message", "HEAD", "-m", "feat(auth): add login"}},
		{name: "reedit", args: []string{"commit", "--reedit-message", "HEAD", "-m", "feat(auth): add login"}},
		{name: "template", args: []string{"commit", "--template", "msg.tmpl", "-m", "feat(auth): add login"}},
		{name: "no-verify", args: []string{"commit", "--no-verify", "-m", "feat(auth): add login"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := executeRoot(t, tt.args, "")
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			if !errors.As(err, new(ExitError)) {
				t.Fatalf("expected ExitError, got %v", err)
			}
			if !strings.Contains(stderr, "unsupported") {
				t.Fatalf("stderr = %q, want unsupported message", stderr)
			}
		})
	}
}

func TestCommitCommand_RemovedFixFlagFails(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeRoot(t, []string{"commit", "--fix", "-m", "feat(auth): add login"}, "")
	if err == nil {
		t.Fatal("expected removed --fix flag to fail")
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(err.Error(), "unknown flag: --fix") {
		t.Fatalf("err = %v, want unknown flag error", err)
	}
}

func executeRoot(t *testing.T, args []string, stdin string) (string, string, error) {
	t.Helper()

	cmd := NewRootCmd("dev")
	cmd.SetArgs(args)
	if stdin != "" {
		cmd.SetIn(strings.NewReader(stdin))
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
