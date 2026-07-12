package cmd

import "testing"

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	t.Run("when a version is injected, resolving version uses it", func(t *testing.T) {
		t.Parallel()

		// Arrange
		injected := "v1.2.3"

		// Act
		version := resolveVersion(injected, "v9.9.9")

		// Assert
		if version != injected {
			t.Fatalf("resolveVersion() = %q, want %q", version, injected)
		}
	})

	t.Run("when no version is injected, resolving version uses the module version", func(t *testing.T) {
		t.Parallel()

		// Arrange
		moduleVersion := "v1.2.3"

		// Act
		version := resolveVersion("dev", moduleVersion)

		// Assert
		if version != moduleVersion {
			t.Fatalf("resolveVersion() = %q, want %q", version, moduleVersion)
		}
	})

	t.Run("when build info is unavailable, resolving version remains dev", func(t *testing.T) {
		t.Parallel()

		// Arrange
		moduleVersion := "(devel)"

		// Act
		version := resolveVersion("dev", moduleVersion)

		// Assert
		if version != "dev" {
			t.Fatalf("resolveVersion() = %q, want dev", version)
		}
	})
}
