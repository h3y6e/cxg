package recall

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/h3y6e/cxg/internal/core/message"
)

var (
	ErrNotGitLog = errors.New("recall: input is not plain git log output")
	ErrQuery     = errors.New("recall: query must be <scope> or <action>(<scope>)")

	actionQueryPattern = regexp.MustCompile(`^([^()]+)\(([^()]+)\)$`)
)

type Query struct {
	Type  string
	Scope string
}

type Commit struct {
	Hash    string
	Subject string
	Actions []message.ActionLine
}

func ParseQuery(value string) (Query, error) {
	if match := actionQueryPattern.FindStringSubmatch(value); match != nil {
		if !slices.Contains(message.ActionTypes, match[1]) {
			return Query{}, fmt.Errorf("recall: unknown action %q, want one of %s", match[1], strings.Join(message.ActionTypes, ", "))
		}

		return Query{Type: match[1], Scope: match[2]}, nil
	}

	if strings.ContainsAny(value, "()") {
		return Query{}, ErrQuery
	}

	return Query{Scope: value}, nil
}

func Recall(log io.Reader, query Query) ([]Commit, error) {
	commits, err := parseLog(log)
	if err != nil {
		return nil, err
	}

	result := []Commit{}

	for _, commit := range commits {
		commit.Actions = slices.DeleteFunc(commit.Actions, func(action message.ActionLine) bool {
			return !query.matches(action)
		})
		if len(commit.Actions) > 0 {
			result = append(result, commit)
		}
	}

	return result, nil
}

func (q Query) matches(action message.ActionLine) bool {
	return (q.Type == "" || q.Type == action.Type) && strings.HasPrefix(action.Scope, q.Scope)
}

type logSection int

const (
	header logSection = iota
	body
	rest
)

func parseLog(log io.Reader) ([]Commit, error) {
	commits := []Commit{}
	section := rest

	content, err := io.ReadAll(log)
	if err != nil {
		return nil, fmt.Errorf("reading git log: %w", err)
	}

	for line := range strings.Lines(string(content)) {
		line = strings.TrimSuffix(line, "\n")

		if hash, ok := strings.CutPrefix(line, "commit "); ok {
			fields := strings.Fields(hash)
			if len(fields) == 0 {
				return nil, ErrNotGitLog
			}

			commits = append(commits, Commit{Hash: fields[0]})
			section = header

			continue
		}

		if len(commits) == 0 {
			if strings.TrimSpace(line) != "" {
				return nil, ErrNotGitLog
			}

			continue
		}

		commit := &commits[len(commits)-1]

		switch {
		case section == header && line == "":
			section = body
		case section != body || line == "":
		case strings.HasPrefix(line, "    "):
			commit.add(strings.TrimPrefix(line, "    "))
		default:
			section = rest
		}
	}

	return commits, nil
}

func (c *Commit) add(line string) {
	if c.Subject == "" {
		c.Subject = line
		return
	}

	if action, ok := message.ParseActionLine(line); ok {
		c.Actions = append(c.Actions, action)
	}
}
