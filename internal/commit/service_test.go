package commit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepare_PrefersMessageFlagsOverStdinAndAppliesFix(t *testing.T) {
	t.Parallel()

	prepared, err := Prepare(PrepareRequest{
		Messages: []string{"feat(auth): add login.\nintent(auth): support social login"},
		Stdin:    strings.NewReader("fix(auth): from stdin"),
		HasStdin: true,
		Fix:      true,
	})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if len(prepared.Errors) != 0 {
		t.Fatalf("Prepare() errors = %#v, want none", prepared.Errors)
	}

	want := "feat(auth): add login\n\nintent(auth): support social login"
	if prepared.Message != want {
		t.Fatalf("Message = %q, want %q", prepared.Message, want)
	}
}

func TestPrepare_PrefersStdinOverFilePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	if err := os.WriteFile(path, []byte("feat(auth): from file\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	prepared, err := Prepare(PrepareRequest{
		FilePath: path,
		Stdin:    strings.NewReader("feat(auth): from stdin"),
		HasStdin: true,
	})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if len(prepared.Errors) != 0 {
		t.Fatalf("Prepare() errors = %#v, want none", prepared.Errors)
	}
	if prepared.Message != "feat(auth): from stdin" {
		t.Fatalf("Message = %q, want stdin content", prepared.Message)
	}
}

func TestRun_CommitsValidatedMessage(t *testing.T) {
	t.Parallel()

	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	result, err := Run(context.Background(), Request{
		Dir: repo,
		PrepareRequest: PrepareRequest{
			Messages: []string{"feat(auth): add login"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Valid {
		t.Fatalf("Valid = %v, want true", result.Valid)
	}
	if !result.Committed {
		t.Fatalf("Committed = %v, want true", result.Committed)
	}
	if result.Message != "feat(auth): add login" {
		t.Fatalf("Message = %q, want %q", result.Message, "feat(auth): add login")
	}
	if result.GitError != nil {
		t.Fatalf("GitError = %#v, want nil", result.GitError)
	}

	message := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
	if message != "feat(auth): add login" {
		t.Fatalf("git log message = %q, want %q", message, "feat(auth): add login")
	}
}

func TestRun_SkipsGitWhenValidationFails(t *testing.T) {
	t.Parallel()

	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	result, err := Run(context.Background(), Request{
		Dir: repo,
		PrepareRequest: PrepareRequest{
			Messages: []string{"bad message"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Valid {
		t.Fatalf("Valid = %v, want false", result.Valid)
	}
	if result.Committed {
		t.Fatalf("Committed = %v, want false", result.Committed)
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "invalid-subject" {
		t.Fatalf("Errors = %#v, want one invalid-subject error", result.Errors)
	}

	count := strings.TrimSpace(runGitAllowFailure(t, repo, "rev-list", "--count", "--all"))
	if count != "0" {
		t.Fatalf("commit count = %q, want %q", count, "0")
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")

	return repo
}

func stageFile(t *testing.T, repo string, name string, content string) {
	t.Helper()

	path := filepath.Join(repo, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	runGit(t, repo, "add", name)
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func runGitAllowFailure(t *testing.T, repo string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return string(out)
		}
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}
