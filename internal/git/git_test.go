package git

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestCommand_UsesProvidedDirectory(t *testing.T) {
	t.Parallel()

	repo := initGitRepo(t)

	cmd, err := Command(context.Background(), repo, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("Command() error = %v", err)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CombinedOutput() error = %v, output = %s", err, out)
	}
	if string(out) != "true\n" {
		t.Fatalf("output = %q, want %q", string(out), "true\n")
	}
}

func TestIsRepository(t *testing.T) {
	t.Parallel()

	repo := initGitRepo(t)
	if !IsRepository(context.Background(), repo) {
		t.Fatal("expected repository directory to be detected")
	}
	if IsRepository(context.Background(), t.TempDir()) {
		t.Fatal("expected non-repository directory to be rejected")
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

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}
