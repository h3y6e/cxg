package message

import "testing"

func TestFix_NormalizesBodyWhitespaceAndSubjectPeriod(t *testing.T) {
	t.Parallel()

	got := Fix("feat(auth): add login.  \n\n\n  intent(auth): support social login  \n\n  decision(auth): keep oauth optional\t")
	want := "feat(auth): add login\n\nintent(auth): support social login\ndecision(auth): keep oauth optional"
	if got != want {
		t.Fatalf("Fix() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestFix_PreservesTrailerBlock(t *testing.T) {
	t.Parallel()

	got := Fix("feat(auth): add login.\nintent(auth): support social login\n\nCo-authored-by: Alice <alice@example.com>  ")
	want := "feat(auth): add login\n\nintent(auth): support social login\n\nCo-authored-by: Alice <alice@example.com>"
	if got != want {
		t.Fatalf("Fix() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
