package lint

import "testing"

func TestValidate_AcceptsSubjectOnlyMessage(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth): add login"

	// Act
	violations := validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("validate() violations = %#v, want empty", violations)
	}
}

func TestValidate_AcceptsBreakingChangeSubject(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "feat(auth)!: add login"

	// Act
	violations := validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("validate() violations = %#v, want empty", violations)
	}
}

func TestValidate_RejectsInvalidSubjectFormat(t *testing.T) {
	t.Parallel()

	// Arrange
	value := "bad message"

	// Act
	violations := validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("validate() violations = %#v, want one violation", violations)
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
	violations := validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("validate() violations = %#v, want one violation", violations)
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
	violations := validate(value)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("validate() violations = %#v, want one violation", violations)
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
	violations := validate(value)

	// Assert
	if len(violations) != 0 {
		t.Fatalf("validate() violations = %#v, want empty", violations)
	}
}
