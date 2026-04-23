package recall

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const (
	commitDelimiter = "---COMMIT_END---"
	commitFormat    = "%H%n%s%n%b" + commitDelimiter
)

// QueryRange returns commits in a range (e.g., "main..HEAD").
func QueryRange(ctx context.Context, dir, rangeSpec string, limit int) ([]CommitEntry, error) {
	args := []string{"log", rangeSpec, "--format=" + commitFormat}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}
	out, err := runGit(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	return parseCommitLog(out), nil
}

// QueryRecent returns the most recent N commits from the current branch.
func QueryRecent(ctx context.Context, dir string, limit int) ([]CommitEntry, error) {
	args := []string{"log", fmt.Sprintf("-%d", limit), "--format=" + commitFormat}
	out, err := runGit(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	return parseCommitLog(out), nil
}

// QueryScopeAll searches all branches for commits with action lines matching a scope prefix.
func QueryScopeAll(ctx context.Context, dir, scope string) ([]CommitEntry, error) {
	pattern := `^(intent|decision|rejected|constraint|learned)\(` + regexp.QuoteMeta(scope)
	args := []string{"log", "--all", "--extended-regexp", "--grep=" + pattern, "--format=" + commitFormat}
	out, err := runGit(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	return parseCommitLog(out), nil
}

// QueryActionScopeAll searches all branches for commits with a specific action(scope) prefix.
func QueryActionScopeAll(ctx context.Context, dir, actionScope string) ([]CommitEntry, error) {
	args := []string{"log", "--all", "--fixed-strings", "--grep=" + actionScope, "--format=" + commitFormat}
	out, err := runGit(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	return parseCommitLog(out), nil
}

func parseCommitLog(output string) []CommitEntry {
	if output == "" {
		return nil
	}

	blocks := strings.Split(output, commitDelimiter)
	var entries []CommitEntry
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		lines := strings.SplitN(block, "\n", 3)
		if len(lines) < 2 {
			continue
		}

		sha := lines[0]
		subject := lines[1]

		var bodyLines []string
		if len(lines) == 3 {
			for _, line := range strings.Split(lines[2], "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					bodyLines = append(bodyLines, trimmed)
				}
			}
		}

		entry := CommitEntry{
			SHA:         sha,
			Subject:     subject,
			ActionLines: ExtractActionLines(bodyLines, sha, subject),
		}
		entries = append(entries, entry)
	}
	return entries
}
