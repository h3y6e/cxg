package cmd

import (
	"bytes"
	"encoding/json"
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

	commitMessage := pipeLintToGitCommit(t, repo, binary, "lint", "-m", "feat(auth): add login")
	if commitMessage != "feat(auth): add login" {
		t.Fatalf("commit message = %q, want %q", commitMessage, "feat(auth): add login")
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

func TestCommitCommand_CreatesCommitFromStagedChange(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "-m", "feat(auth): add login")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	message := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
	if message != "feat(auth): add login" {
		t.Fatalf("commit message = %q, want %q", message, "feat(auth): add login")
	}
}

func TestCommitCommand_ValidationFailureDoesNotCreateCommit(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "-m", "bad message")
	if err == nil {
		t.Fatal("expected commit command to fail")
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "invalid-subject") {
		t.Fatalf("stderr = %q, want invalid-subject", stderr)
	}

	count := strings.TrimSpace(runGit(t, repo, "rev-list", "--count", "--all"))
	if count != "0" {
		t.Fatalf("commit count = %q, want %q", count, "0")
	}
}

func TestCommitCommand_FixesMessageByDefault(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "-m", "feat(auth): add login.\nintent(auth): support social login")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	want := "feat(auth): add login\n\nintent(auth): support social login"
	message := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
	if message != want {
		t.Fatalf("commit message = %q, want %q", message, want)
	}
}

func TestCommitCommand_AllowEmptyCreatesCommit(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--allow-empty", "-m", "chore(repo): initialize")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	count := strings.TrimSpace(runGit(t, repo, "rev-list", "--count", "--all"))
	if count != "1" {
		t.Fatalf("commit count = %q, want %q", count, "1")
	}
}

func TestCommitCommand_AllStagesTrackedChanges(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	writeCommittedFile(t, repo, "README.md", "before\n", "docs(repo): seed readme")
	stageFile(t, repo, "README.md", "after\n")
	runGit(t, repo, "reset", "HEAD", "README.md")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--all", "-m", "docs(readme): refresh usage")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	content := runGit(t, repo, "show", "HEAD:README.md")
	if content != "after\n" {
		t.Fatalf("committed file = %q, want %q", content, "after\n")
	}
}

func TestCommitCommand_AmendUpdatesPreviousCommitMessage(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	writeCommittedFile(t, repo, "README.md", "before\n", "docs(repo): seed readme")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--amend", "--allow-empty", "-m", "docs(readme): refresh usage")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	count := strings.TrimSpace(runGit(t, repo, "rev-list", "--count", "--all"))
	if count != "1" {
		t.Fatalf("commit count = %q, want %q", count, "1")
	}
	message := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
	if message != "docs(readme): refresh usage" {
		t.Fatalf("commit message = %q, want %q", message, "docs(readme): refresh usage")
	}
}

func TestCommitCommand_SignoffAuthorAndDate(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(
		t,
		repo,
		binary,
		"commit",
		"--signoff",
		"--author", "Author Name <author@example.com>",
		"--date", "2026-04-06T12:34:56Z",
		"-m", "feat(auth): add login",
	)
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}

	author := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%an <%ae>"))
	if author != "Author Name <author@example.com>" {
		t.Fatalf("author = %q, want %q", author, "Author Name <author@example.com>")
	}
	date := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%aI"))
	if date != "2026-04-06T12:34:56Z" {
		t.Fatalf("date = %q, want %q", date, "2026-04-06T12:34:56Z")
	}
	message := strings.TrimSpace(runGit(t, repo, "log", "-1", "--pretty=%B"))
	if !strings.Contains(message, "Signed-off-by: Test User <test@example.com>") {
		t.Fatalf("message = %q, want Signed-off-by trailer", message)
	}
}

func TestCommitCommand_HookFailureIsSurfaced(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho blocked >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "-m", "feat(auth): add login")
	if err == nil {
		t.Fatal("expected hook failure")
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "blocked") {
		t.Fatalf("stderr = %q, want hook output", stderr)
	}
}

func TestCommitCommand_JSONSuccess(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--json", "-m", "feat(auth): add login")
	if err != nil {
		t.Fatalf("runCxg() error = %v, stdout = %s, stderr = %s", err, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got struct {
		Valid     bool   `json:"valid"`
		Committed bool   `json:"committed"`
		Message   string `json:"message"`
		Errors    []any  `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !got.Valid || !got.Committed {
		t.Fatalf("got = %#v, want valid and committed true", got)
	}
	if got.Message != "feat(auth): add login" {
		t.Fatalf("Message = %q, want %q", got.Message, "feat(auth): add login")
	}
}

func TestCommitCommand_JSONValidationFailure(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)
	stageFile(t, repo, "README.md", "hello\n")

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--json", "-m", "bad message")
	if err == nil {
		t.Fatal("expected validation failure")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got struct {
		Valid     bool `json:"valid"`
		Committed bool `json:"committed"`
		Errors    []struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.Valid || got.Committed {
		t.Fatalf("got = %#v, want valid=false committed=false", got)
	}
	if len(got.Errors) != 1 || got.Errors[0].Code != "invalid-subject" {
		t.Fatalf("Errors = %#v, want one invalid-subject error", got.Errors)
	}
}

func TestCommitCommand_JSONGitFailure(t *testing.T) {
	t.Parallel()

	binary := buildCxgBinary(t)
	repo := initGitRepo(t)

	stdout, stderr, err := runCxg(t, repo, binary, "commit", "--json", "-m", "feat(auth): add login")
	if err == nil {
		t.Fatal("expected git failure")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got struct {
		Valid     bool `json:"valid"`
		Committed bool `json:"committed"`
		GitError  *struct {
			ExitCode int    `json:"exitCode"`
			Stderr   string `json:"stderr"`
		} `json:"gitError"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !got.Valid || got.Committed {
		t.Fatalf("got = %#v, want valid=true committed=false", got)
	}
	if got.GitError == nil || got.GitError.ExitCode == 0 {
		t.Fatalf("GitError = %#v, want non-zero exit code", got.GitError)
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

func writeCommittedFile(t *testing.T, repo string, name string, content string, message string) {
	t.Helper()

	stageFile(t, repo, name, content)
	runGit(t, repo, "commit", "-m", message)
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

func runCxg(t *testing.T, repo string, binary string, args ...string) (string, string, error) {
	t.Helper()

	command := exec.Command(binary, args...)
	command.Dir = repo

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.String(), stderr.String(), err
}
