package message

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_AcceptsSubjectOnlyMessage(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): add login"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("Validate() violations = %#v, want empty", violations)
	}
}

func TestValidate_AcceptsBreakingChangeSubject(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth)!: add login"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("Validate() violations = %#v, want empty", violations)
	}
}

func TestValidate_RejectsInvalidSubjectFormat(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "bad message"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Validate() violations = %#v, want one violation", violations)
	}

	if violations[0].Line != 1 {
		t.Fatalf("Line = %d, want 1", violations[0].Line)
	}

	if violations[0].Code != "invalid-subject" {
		t.Fatalf("Code = %q, want %q", violations[0].Code, "invalid-subject")
	}
}

func TestValidate_RejectsLongSubject(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): this subject is intentionally made much longer than seventy-two characters for validation"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Validate() violations = %#v, want one violation", violations)
	}

	if violations[0].Code != "subject-too-long" {
		t.Fatalf("Code = %q, want %q", violations[0].Code, "subject-too-long")
	}
}

func TestValidate_RejectsInvalidActionLine(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): add login\n\nnote(auth): unsupported action"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Validate() violations = %#v, want one violation", violations)
	}

	if violations[0].Line != 3 {
		t.Fatalf("Line = %d, want 3", violations[0].Line)
	}

	if violations[0].Code != "invalid-action-format" {
		t.Fatalf("Code = %q, want %q", violations[0].Code, "invalid-action-format")
	}
}

func TestValidate_IgnoresTrailers(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): add login\n\nintent(auth): support social login\n\nCo-authored-by: Alice <alice@example.com>"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("Validate() violations = %#v, want empty", violations)
	}
}

func TestValidate_ReportsPhysicalLineAfterExtraBlankLines(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): add login\n\n\nnote(auth): unsupported action"

	// Act
	violations := Validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Validate() violations = %#v, want one violation", violations)
	}

	if violations[0].Line != 4 {
		t.Fatalf("Line = %d, want 4", violations[0].Line)
	}
}

func TestValidate_AcceptsContextualCommitSkillExamples(t *testing.T) {
	t.Parallel()

	// Arrange
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "specs", "cli", "research", "contextual-commit", "SKILL.md"))
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
		if !strings.Contains(string(content), example) {
			t.Fatalf("skill file is missing expected example:\n%s", example)
		}

		// Act
		violations := Validate(example)

		// Assert
		if len(violations) != 0 {
			t.Fatalf("Validate() violations = %#v, want empty\nexample:\n%s", violations, example)
		}
	}
}
