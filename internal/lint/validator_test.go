package lint

import "testing"

func TestValidate_AcceptsSubjectOnlyMessage(t *testing.T) {
	t.Parallel()

	errors := Validate("feat(auth): add login")
	if len(errors) != 0 {
		t.Fatalf("Validate() returned errors: %#v", errors)
	}
}

func TestValidate_AcceptsBreakingChangeSubject(t *testing.T) {
	t.Parallel()

	errors := Validate("feat(auth)!: add login")
	if len(errors) != 0 {
		t.Fatalf("Validate() returned errors: %#v", errors)
	}
}

func TestValidate_RejectsInvalidSubjectFormat(t *testing.T) {
	t.Parallel()

	errors := Validate("bad message")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %#v", errors)
	}

	if errors[0].Line != 1 {
		t.Fatalf("Line = %d, want 1", errors[0].Line)
	}
	if errors[0].Code != "invalid-subject" {
		t.Fatalf("Code = %q, want %q", errors[0].Code, "invalid-subject")
	}
}

func TestValidate_RejectsLongSubject(t *testing.T) {
	t.Parallel()

	errors := Validate("feat(auth): this subject is intentionally made much longer than seventy-two characters for validation")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %#v", errors)
	}

	if errors[0].Code != "subject-too-long" {
		t.Fatalf("Code = %q, want %q", errors[0].Code, "subject-too-long")
	}
}

func TestValidate_RejectsInvalidActionLine(t *testing.T) {
	t.Parallel()

	errors := Validate("feat(auth): add login\n\nnote(auth): unsupported action")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %#v", errors)
	}

	if errors[0].Line != 3 {
		t.Fatalf("Line = %d, want 3", errors[0].Line)
	}
	if errors[0].Code != "invalid-action-format" {
		t.Fatalf("Code = %q, want %q", errors[0].Code, "invalid-action-format")
	}
}

func TestValidate_IgnoresTrailers(t *testing.T) {
	t.Parallel()

	errors := Validate("feat(auth): add login\n\nintent(auth): support social login\n\nCo-authored-by: Alice <alice@example.com>")
	if len(errors) != 0 {
		t.Fatalf("Validate() returned errors: %#v", errors)
	}
}
