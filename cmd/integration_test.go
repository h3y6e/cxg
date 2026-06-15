package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	cxglint "github.com/h3y6e/cxg/internal/lint"
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

	command := exec.Command("git", "commit", "--allow-empty", "-m", "bad message")
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

func TestContextualCommitSkillExamplesValidate(t *testing.T) {
	t.Parallel()

	skillPath := filepath.Join("..", "specs", "cli", "research", "contextual-commit", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	examples := []string{
		"fix(button): correct alignment on mobile viewport",
		"feat(notifications): add email digest for weekly summaries\n\nintent(notifications): users want batch notifications instead of per-event emails\ndecision(digest-schedule): weekly on Monday 9am — matches user research feedback\nconstraint(email-provider): SendGrid batch API limited to 1000 recipients per call",
		"refactor(payments): migrate from single to multi-currency support\n\nintent(payments): enterprise customers need EUR and GBP alongside USD\nintent(payment-architecture): must be backward compatible, existing USD flows unchanged\ndecision(currency-handling): per-transaction currency over account-level default\nrejected(currency-handling): account-level default too limiting for marketplace sellers\nrejected(money-library): accounting.js — lacks sub-unit arithmetic, using currency.js instead\nconstraint(stripe-integration): Stripe requires currency at PaymentIntent creation, cannot change after\nconstraint(database-migration): existing amount columns need companion currency columns, not replacement\nlearned(stripe-multicurrency): presentment currency vs settlement currency are different Stripe concepts\nlearned(exchange-rates): Stripe handles conversion, we should NOT store our own rates",
		"refactor(auth): switch from session-based to JWT tokens\n\nintent(auth): original session approach incompatible with redis cluster setup\nrejected(auth-sessions): redis cluster doesn't support session stickiness needed by passport sessions\ndecision(auth-tokens): JWT with short expiry + refresh token pattern\nlearned(redis-cluster): session affinity requires sticky sessions at load balancer level — too invasive",
	}

	for _, example := range examples {
		if !bytes.Contains(content, []byte(example)) {
			t.Fatalf("skill file is missing expected example:\n%s", example)
		}

		errors := cxglint.Validate(example)
		if len(errors) != 0 {
			t.Fatalf("example should validate without errors, got %#v\nexample:\n%s", errors, example)
		}
	}
}

func buildCxgBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "cxg")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = ".."

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

func pipeLintToGitCommit(t *testing.T, repo string, binary string, args ...string) string {
	t.Helper()

	lintCommand := exec.Command(binary, args...)
	lintCommand.Dir = repo

	pipe, err := lintCommand.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe() error = %v", err)
	}

	var lintStderr bytes.Buffer
	lintCommand.Stderr = &lintStderr

	commitCommand := exec.Command("git", "commit", "--allow-empty", "-F", "-")
	commitCommand.Dir = repo
	commitCommand.Stdin = pipe

	var commitStderr bytes.Buffer
	commitCommand.Stderr = &commitStderr

	if err := lintCommand.Start(); err != nil {
		t.Fatalf("lint Start() error = %v", err)
	}
	if err := commitCommand.Run(); err != nil {
		t.Fatalf("git commit error = %v, lint stderr = %s, git stderr = %s", err, lintStderr.String(), commitStderr.String())
	}
	if err := lintCommand.Wait(); err != nil {
		t.Fatalf("lint Wait() error = %v, lint stderr = %s", err, lintStderr.String())
	}

	return strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = repo

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		t.Fatalf("git %v error = %v, stderr = %s", args, err, stderr.String())
	}

	return stdout.String()
}
