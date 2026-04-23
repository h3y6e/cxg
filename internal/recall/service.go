package recall

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyRepository = errors.New("repository has no commits")

// Request specifies what recall should do.
type Request struct {
	Dir  string
	Mode Mode
	// Query is the raw argument: scope or action(scope)
	Query string
	// For action+scope mode, parsed components
	ActionType string
	Scope      string
}

// ParseArgument detects the recall mode from the CLI argument.
func ParseArgument(arg string) Request {
	if arg == "" {
		return Request{Mode: ModeDefault}
	}
	if idx := strings.Index(arg, "("); idx > 0 && strings.HasSuffix(arg, ")") {
		return Request{
			Mode:       ModeActionScope,
			Query:      arg,
			ActionType: arg[:idx],
			Scope:      arg[idx+1 : len(arg)-1],
		}
	}
	return Request{Mode: ModeScope, Query: arg, Scope: arg}
}

// Run executes a recall request and returns the result.
func Run(ctx context.Context, req Request) (RecallResult, error) {
	switch req.Mode {
	case ModeScope:
		return runScopeQuery(ctx, req)
	case ModeActionScope:
		return runActionScopeQuery(ctx, req)
	default:
		return runDefault(ctx, req)
	}
}

func runDefault(ctx context.Context, req Request) (RecallResult, error) {
	if _, err := runGit(ctx, req.Dir, "rev-parse", "HEAD"); err != nil {
		return RecallResult{}, ErrEmptyRepository
	}

	branch, err := DetectBranchState(ctx, req.Dir)
	if err != nil {
		return RecallResult{}, err
	}

	var commits []CommitEntry
	switch {
	case branch.IsDefault:
		commits, err = QueryRecent(ctx, req.Dir, 20)
	case branch.CommitsAhead > 0:
		commits, err = QueryRange(ctx, req.Dir, branch.Base+"..HEAD", 0)
	default:
		// Feature branch has no commits yet — read recent commits from the base
		// branch so the result reflects its current state, not a stale snapshot.
		commits, err = QueryRange(ctx, req.Dir, branch.Base, 10)
	}
	if err != nil {
		return RecallResult{}, err
	}

	return RecallResult{
		Mode:        ModeDefault,
		Branch:      branch,
		Commits:     commits,
		ActionLines: flattenActionLines(commits),
	}, nil
}

func runScopeQuery(ctx context.Context, req Request) (RecallResult, error) {
	commits, err := QueryScopeAll(ctx, req.Dir, req.Scope)
	if err != nil {
		return RecallResult{}, err
	}

	var filtered []ActionLine
	scopeSet := make(map[string]bool)
	for _, c := range commits {
		for _, al := range c.ActionLines {
			if strings.HasPrefix(al.Scope, req.Scope) {
				filtered = append(filtered, al)
				scopeSet[al.Scope] = true
			}
		}
	}

	scopes := make([]string, 0, len(scopeSet))
	for s := range scopeSet {
		scopes = append(scopes, s)
	}

	return RecallResult{
		Mode:        ModeScope,
		Query:       req.Query,
		Commits:     commits,
		ActionLines: filtered,
		Scopes:      scopes,
	}, nil
}

func runActionScopeQuery(ctx context.Context, req Request) (RecallResult, error) {
	commits, err := QueryActionScopeAll(ctx, req.Dir, req.ActionType+"("+req.Scope)
	if err != nil {
		return RecallResult{}, err
	}

	var filtered []ActionLine
	for _, c := range commits {
		for _, al := range c.ActionLines {
			if al.Type == req.ActionType && strings.HasPrefix(al.Scope, req.Scope) {
				filtered = append(filtered, al)
			}
		}
	}

	return RecallResult{
		Mode:        ModeActionScope,
		Query:       req.Query,
		Commits:     commits,
		ActionLines: filtered,
	}, nil
}

func flattenActionLines(commits []CommitEntry) []ActionLine {
	var result []ActionLine
	for _, c := range commits {
		result = append(result, c.ActionLines...)
	}
	return result
}
