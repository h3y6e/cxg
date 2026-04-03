package message

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_UsesMessagesBeforeOtherInputs(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "COMMIT_EDITMSG")
	if err := os.WriteFile(filePath, []byte("feat(file): from file\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got, err := Resolve(Input{
		Messages: []string{"feat(auth): add login"},
		FilePath: filePath,
		Stdin:    strings.NewReader("feat(stdin): from stdin"),
		HasStdin: true,
	})
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if got != "feat(auth): add login" {
		t.Fatalf("Resolve() = %q, want %q", got, "feat(auth): add login")
	}
}

func TestResolve_UsesStdinBeforeFilePath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "COMMIT_EDITMSG")
	if err := os.WriteFile(filePath, []byte("feat(file): from file\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got, err := Resolve(Input{
		FilePath: filePath,
		Stdin:    strings.NewReader("feat(stdin): from stdin"),
		HasStdin: true,
	})
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if got != "feat(stdin): from stdin" {
		t.Fatalf("Resolve() = %q, want %q", got, "feat(stdin): from stdin")
	}
}

func TestResolve_JoinsMessagesAndAppendsTrailers(t *testing.T) {
	got, err := Resolve(Input{
		Messages: []string{
			"feat(auth): add login",
			"intent(auth): support social login",
		},
		Trailers: []string{
			"Co-authored-by: Alice <alice@example.com>",
		},
	})
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	want := strings.Join([]string{
		"feat(auth): add login",
		"",
		"intent(auth): support social login",
		"",
		"Co-authored-by: Alice <alice@example.com>",
	}, "\n")
	if got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolve_ReadsFilePathWhenNoOtherInputExists(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "COMMIT_EDITMSG")
	if err := os.WriteFile(filePath, []byte("feat(auth): from file\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got, err := Resolve(Input{FilePath: filePath})
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if got != "feat(auth): from file" {
		t.Fatalf("Resolve() = %q, want %q", got, "feat(auth): from file")
	}
}
