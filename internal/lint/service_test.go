package lint_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/lint"
)

func TestServiceRun(t *testing.T) {
	t.Parallel()

	t.Run("when message flags need fixing, running lint returns the normalized valid message", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := lint.New(func(string) ([]byte, error) {
			t.Fatal("read file must not be called")
			return nil, nil
		})
		input := lint.Input{
			Messages: []string{
				"feat(auth): add login.",
				"  intent(auth): support social login",
			},
			Trailers: []string{"Co-authored-by: Alice <alice@example.com>"},
			Fix:      true,
		}

		// Act
		result, err := service.Run(input)

		// Assert
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if len(result.Violations) != 0 {
			t.Fatalf("Run() violations = %#v, want empty", result.Violations)
		}
		expected := "feat(auth): add login\n\nintent(auth): support social login\n\nCo-authored-by: Alice <alice@example.com>"
		if result.Message != expected {
			t.Fatalf("Run() message = %q, want %q", result.Message, expected)
		}
	})

	t.Run("when stdin and a file are provided, running lint uses stdin without reading the file", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := lint.New(func(string) ([]byte, error) {
			t.Fatal("read file must not be called")
			return nil, nil
		})
		input := lint.Input{
			Stdin:    strings.NewReader("feat(stdin): from stdin\n"),
			FilePath: "COMMIT_EDITMSG",
		}

		// Act
		result, err := service.Run(input)

		// Assert
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.Message != "feat(stdin): from stdin" {
			t.Fatalf("Run() message = %q, want stdin message", result.Message)
		}
	})

	t.Run("when only a file is provided, running lint reads it through the file port", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := lint.New(func(path string) ([]byte, error) {
			if path != "COMMIT_EDITMSG" {
				t.Fatalf("read file path = %q, want COMMIT_EDITMSG", path)
			}
			return []byte("feat(file): from file\n"), nil
		})

		// Act
		result, err := service.Run(lint.Input{FilePath: "COMMIT_EDITMSG"})

		// Assert
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.Message != "feat(file): from file" {
			t.Fatalf("Run() message = %q, want file message", result.Message)
		}
	})

	t.Run("when no input is provided, running lint returns ErrNoInput", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := lint.New(func(string) ([]byte, error) {
			t.Fatal("read file must not be called")
			return nil, nil
		})

		// Act
		_, err := service.Run(lint.Input{})

		// Assert
		if !errors.Is(err, lint.ErrNoInput) {
			t.Fatalf("Run() error = %v, want ErrNoInput", err)
		}
	})

	t.Run("when reading a file fails, running lint wraps the file path and cause", func(t *testing.T) {
		t.Parallel()

		// Arrange
		expectedCause := errors.New("permission denied")
		service := lint.New(func(string) ([]byte, error) {
			return nil, expectedCause
		})

		// Act
		_, err := service.Run(lint.Input{FilePath: "COMMIT_EDITMSG"})

		// Assert
		if !errors.Is(err, expectedCause) {
			t.Fatalf("Run() error = %v, want wrapped cause", err)
		}
		expected := fmt.Sprintf("reading file %q: permission denied", "COMMIT_EDITMSG")
		if err.Error() != expected {
			t.Fatalf("Run() error = %q, want %q", err, expected)
		}
	})

	t.Run("when an invalid action follows extra blank lines, running lint reports its physical line", func(t *testing.T) {
		t.Parallel()

		// Arrange
		service := lint.New(func(string) ([]byte, error) {
			t.Fatal("read file must not be called")
			return nil, nil
		})
		input := lint.Input{Messages: []string{"feat(auth): add login\n\n\nnote(auth): unsupported action"}}

		// Act
		result, err := service.Run(input)

		// Assert
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if len(result.Violations) != 1 {
			t.Fatalf("Run() violations = %#v, want one violation", result.Violations)
		}
		if result.Violations[0].Line != 4 {
			t.Fatalf("Run() violation line = %d, want 4", result.Violations[0].Line)
		}
	})
}
