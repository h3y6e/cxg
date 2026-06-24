package recall

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	internalgit "github.com/h3y6e/cxg/internal/git"
)

func DetectBranchState(ctx context.Context, dir string) (BranchState, error) {
	current, err := runGit(ctx, dir, "branch", "--show-current")
	if err != nil {
		return BranchState{}, err
	}

	defaultBranch, err := resolveDefaultBranch(ctx, dir)
	if err != nil {
		return BranchState{}, err
	}

	isDefault := current == defaultBranch

	baseBranch, err := resolveBaseBranch(ctx, dir, current, defaultBranch)
	if err != nil {
		return BranchState{}, err
	}

	commitsAhead, err := countCommitsAhead(ctx, dir, baseBranch)
	if err != nil {
		return BranchState{}, err
	}

	hasUnstaged := hasOutput(ctx, dir, "diff", "--quiet")
	hasStaged := hasOutput(ctx, dir, "diff", "--cached", "--quiet")

	return BranchState{
		Current:      current,
		Base:         baseBranch,
		CommitsAhead: commitsAhead,
		IsDefault:    isDefault,
		HasUnstaged:  hasUnstaged,
		HasStaged:    hasStaged,
	}, nil
}

func resolveDefaultBranch(ctx context.Context, dir string) (string, error) {
	out, err := runGit(ctx, dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil && out != "" {
		return strings.TrimPrefix(out, "refs/remotes/origin/"), nil
	}
	for _, candidate := range []string{"main", "master"} {
		if _, err := runGit(ctx, dir, "rev-parse", "--verify", "refs/heads/"+candidate); err == nil {
			return candidate, nil
		}
	}
	out, err = runGit(ctx, dir, "branch", "--format=%(refname:short)")
	if err == nil && out != "" {
		first := strings.SplitN(out, "\n", 2)[0]
		if first != "" {
			return first, nil
		}
	}
	return "main", nil
}

func resolveBaseBranch(ctx context.Context, dir, current, defaultBranch string) (string, error) {
	upstream, err := runGit(ctx, dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err == nil && upstream != "" {
		// Strip the remote prefix to get the branch name.
		// If it still equals the current branch, the tracking ref points to
		// the branch itself (e.g. after `git push -u`), not its parent — skip it.
		candidate := strings.TrimPrefix(upstream, "origin/")
		if candidate != current {
			return candidate, nil
		}
	}

	branches, err := runGit(ctx, dir, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err == nil && branches != "" {
		nearest := ""
		minDistance := -1
		for _, branch := range strings.Split(branches, "\n") {
			if branch == "" || branch == current {
				continue
			}
			if !isAncestor(ctx, dir, branch, current) {
				continue
			}
			count, err := countCommitsInRange(ctx, dir, branch+".."+current)
			if err != nil {
				continue
			}
			if minDistance < 0 || count < minDistance {
				minDistance = count
				nearest = branch
			}
		}
		if nearest != "" {
			return nearest, nil
		}
	}

	return defaultBranch, nil
}

// isAncestor returns true if candidate is an ancestor of current.
func isAncestor(ctx context.Context, dir, candidate, current string) bool {
	cmd, err := internalgit.Command(ctx, dir, "merge-base", "--is-ancestor", candidate, current)
	if err != nil {
		return false
	}
	return cmd.Run() == nil
}

func countCommitsAhead(ctx context.Context, dir, baseBranch string) (int, error) {
	return countCommitsInRange(ctx, dir, baseBranch+"..HEAD")
}

func countCommitsInRange(ctx context.Context, dir, rangeSpec string) (int, error) {
	out, err := runGit(ctx, dir, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0, err
	}
	var count int
	if _, err := fmt.Sscanf(out, "%d", &count); err != nil {
		return 0, err
	}
	return count, nil
}

// hasOutput returns true if the git command exits non-zero (i.e., there are changes).
func hasOutput(ctx context.Context, dir string, args ...string) bool {
	cmd, err := internalgit.Command(ctx, dir, args...)
	if err != nil {
		return false
	}
	return cmd.Run() != nil
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd, err := internalgit.Command(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}
