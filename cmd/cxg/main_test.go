package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPipeline_CreatesCommitFromLintOutput(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)

	commitMessage := pipeLintToGitCommit(
		t,
		repo,
		binary,
		"lint",
		"-m", "feat(auth): add login",
		"-m", "intent(auth): support social login",
		"-m", "decision(auth): keep OAuth optional",
	)

	want := "feat(auth): add login\n\nintent(auth): support social login\ndecision(auth): keep OAuth optional"
	if commitMessage != want {
		t.Fatalf("commit message = %q, want %q", commitMessage, want)
	}
}

func TestCommitMsgHook_RejectsInvalidMessage(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)

	hookPath := filepath.Join(repo, ".git", "hooks", "commit-msg")

	hook := "#!/bin/sh\n\"" + binary + "\" lint \"$1\"\n"
	if err := os.WriteFile(hookPath, []byte(hook), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	command := exec.CommandContext(t.Context(), "git", "commit", "--allow-empty", "-m", "bad message")
	command.Dir = repo

	var stderr bytes.Buffer

	command.Stderr = &stderr
	if err := command.Run(); err == nil {
		t.Fatal("expected commit to fail")
	}

	if !strings.Contains(stderr.String(), "invalid-subject") {
		t.Fatalf("stderr = %q, want invalid-subject", stderr.String())
	}
}

func TestFixPipeline_CreatesNormalizedCommit(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)

	commitMessage := pipeLintToGitCommit(
		t,
		repo,
		binary,
		"lint",
		"--fix",
		"-m", "feat(auth): add login.\nintent(auth): support social login",
	)

	want := "feat(auth): add login\n\nintent(auth): support social login"
	if commitMessage != want {
		t.Fatalf("commit message = %q, want %q", commitMessage, want)
	}
}

func TestRecallPipeline_ReadsActionLinesFromGitLog(t *testing.T) {
	t.Parallel()

	// Arrange
	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	runGit(t, repo, "commit", "--allow-empty", "-m", "feat(auth): add login", "-m", "rejected(auth): auth0-sdk session model")
	runGit(t, repo, "commit", "--allow-empty", "-m", "fix(payments): round amounts", "-m", "constraint(payments): integer cents only")

	recallCommand := exec.CommandContext(t.Context(), binary, "recall", "rejected(auth)")
	recallCommand.Dir = repo
	recallCommand.Stdin = strings.NewReader(runGit(t, repo, "log", "--stat"))

	// Act
	output, err := recallCommand.Output()

	// Assert
	if err != nil {
		t.Fatalf("recall error = %v", err)
	}

	hash := runGit(t, repo, "rev-parse", "--short=7", "HEAD~1")

	want := strings.TrimSpace(hash) + " feat(auth): add login\n  rejected(auth): auth0-sdk session model\n"
	if string(output) != want {
		t.Fatalf("recall output = %q, want %q", output, want)
	}
}

func buildCxgBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "cxg")
	command := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")

	var stderr bytes.Buffer

	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("go build error = %v, stderr = %s", err, stderr.String())
	}

	return binary
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "commit.gpgsign", "false")

	return repo
}

func pipeLintToGitCommit(t *testing.T, repo, binary string, args ...string) string {
	t.Helper()

	lintCommand := exec.CommandContext(t.Context(), binary, args...)
	lintCommand.Dir = repo

	pipe, err := lintCommand.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe() error = %v", err)
	}

	var lintStderr bytes.Buffer

	lintCommand.Stderr = &lintStderr

	commitCommand := exec.CommandContext(t.Context(), "git", "commit", "--allow-empty", "-F", "-")
	commitCommand.Dir = repo
	commitCommand.Stdin = pipe

	var commitStderr bytes.Buffer

	commitCommand.Stderr = &commitStderr

	if err := lintCommand.Start(); err != nil {
		t.Fatalf("lint Start() error = %v", err)
	}

	commitErr := commitCommand.Run()
	lintErr := lintCommand.Wait()

	if commitErr != nil {
		t.Fatalf("git commit error = %v, lint stderr = %s, git stderr = %s", commitErr, lintStderr.String(), commitStderr.String())
	}

	if lintErr != nil {
		t.Fatalf("lint Wait() error = %v, lint stderr = %s", lintErr, lintStderr.String())
	}

	return strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()

	command := exec.CommandContext(t.Context(), "git", args...)
	command.Dir = repo

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		t.Fatalf("git %v error = %v, stderr = %s", args, err, stderr.String())
	}

	return stdout.String()
}
