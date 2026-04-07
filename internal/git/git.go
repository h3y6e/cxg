package git

import (
	"context"
	"fmt"
	"os/exec"
)

func Command(ctx context.Context, dir string, args ...string) (*exec.Cmd, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Dir = dir
	return cmd, nil
}

func IsRepository(ctx context.Context, dir string) bool {
	cmd, err := Command(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return false
	}

	return cmd.Run() == nil
}
