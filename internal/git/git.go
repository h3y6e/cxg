package git

import (
	"context"
	"fmt"
	"os/exec"
)

// Command creates an exec.Cmd for running a git command in the given directory.
func Command(ctx context.Context, dir string, args ...string) (*exec.Cmd, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found: %w", err)
	}
	fullArgs := append([]string{"-C", dir}, args...)
	return exec.CommandContext(ctx, gitPath, fullArgs...), nil
}

// IsRepository returns true if dir is inside a git repository.
func IsRepository(ctx context.Context, dir string) bool {
	cmd, err := Command(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}
