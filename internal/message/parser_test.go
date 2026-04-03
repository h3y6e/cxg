package message

import (
	"reflect"
	"testing"
)

func TestParse_SubjectOnlyMessage(t *testing.T) {
	got, err := Parse("feat(auth): add login")
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	want := CommitMessage{
		Subject: "feat(auth): add login",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}

func TestParse_SplitsBodyAndTrailers(t *testing.T) {
	got, err := Parse("feat(auth): add login\n\nintent(auth): support social login\n\ndecision(auth): keep OAuth optional\n\nCo-authored-by: Alice <alice@example.com>")
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	want := CommitMessage{
		Subject: "feat(auth): add login",
		BodyLines: []string{
			"intent(auth): support social login",
			"decision(auth): keep OAuth optional",
		},
		Trailers: []string{
			"Co-authored-by: Alice <alice@example.com>",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}
