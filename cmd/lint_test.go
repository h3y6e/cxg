package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintCommand_WritesValidMessageToStdout(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{
		"lint",
		"-m", "feat(auth): add login",
		"-m", "intent(auth): support social login",
		"-m", "decision(auth): keep OAuth optional",
		"--trailer", "Co-authored-by: Alice <alice@example.com>",
	}, "")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}

	want := strings.Join([]string{
		"feat(auth): add login",
		"",
		"intent(auth): support social login",
		"decision(auth): keep OAuth optional",
		"",
		"Co-authored-by: Alice <alice@example.com>",
	}, "\n")
	if stdout != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\n\ngot:\n%s", want, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestLintCommand_WritesValidationErrorsToStderr(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{"lint", "-m", "bad message"}, "")
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !errors.As(err, new(ExitError)) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if !strings.Contains(stderr, "invalid-subject") {
		t.Fatalf("stderr = %q, want invalid-subject", stderr)
	}
}

func TestLintCommand_AcceptsBreakingChangeSubject(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{"lint", "-m", "feat(auth)!: add login"}, "")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}
	if stdout != "feat(auth)!: add login" {
		t.Fatalf("stdout = %q, want %q", stdout, "feat(auth)!: add login")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestLintCommand_ReadsMessageFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	content := "feat(auth): add login\n\nintent(auth): support social login\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout, stderr, err := executeLint(t, []string{"lint", path}, "")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}

	want := strings.TrimRight(content, "\n")
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestLintCommand_PrefersMessageFlagsOverStdin(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{"lint", "-m", "feat(auth): add login"}, "fix(auth): from stdin")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}

	if stdout != "feat(auth): add login" {
		t.Fatalf("stdout = %q, want message flag content", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestLintCommand_FixNormalizesMessageBeforeValidation(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{
		"lint",
		"--fix",
		"-m", "feat(auth): add login.\nintent(auth): support social login",
	}, "")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}

	want := "feat(auth): add login\n\nintent(auth): support social login"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestLintCommand_FixStillFailsWhenMessageRemainsInvalid(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{
		"lint",
		"--fix",
		"-m", "bad message",
	}, "")
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !errors.As(err, new(ExitError)) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if !strings.Contains(stderr, "invalid-subject") {
		t.Fatalf("stderr = %q, want invalid-subject", stderr)
	}
}

func TestLintCommand_JSONOutputsStructuredSuccess(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{
		"lint",
		"--json",
		"-m", "feat(auth): add login",
	}, "")
	if err != nil {
		t.Fatalf("executeLint() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
		Errors  []any  `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !got.Valid {
		t.Fatalf("Valid = %v, want true", got.Valid)
	}
	if got.Message != "feat(auth): add login" {
		t.Fatalf("Message = %q, want %q", got.Message, "feat(auth): add login")
	}
	if len(got.Errors) != 0 {
		t.Fatalf("Errors = %#v, want empty", got.Errors)
	}
}

func TestLintCommand_JSONOutputsStructuredFailure(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeLint(t, []string{
		"lint",
		"--json",
		"-m", "bad message",
	}, "")
	if !errors.As(err, new(ExitError)) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got struct {
		Valid  bool `json:"valid"`
		Errors []struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got.Valid {
		t.Fatalf("Valid = %v, want false", got.Valid)
	}
	if len(got.Errors) != 1 || got.Errors[0].Code != "invalid-subject" {
		t.Fatalf("Errors = %#v, want one invalid-subject error", got.Errors)
	}
}

func executeLint(t *testing.T, args []string, stdin string) (string, string, error) {
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
