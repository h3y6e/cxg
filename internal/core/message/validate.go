package message

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	subjectTypes      = []string{"feat", "fix", "refactor", "perf", "test", "docs", "style", "build", "ci", "chore", "revert"}
	subjectPattern    = regexp.MustCompile(`^(` + strings.Join(subjectTypes, "|") + `)(\([^)]+\))?(!)?: .+$`)
	ActionTypes       = []string{"intent", "decision", "rejected", "constraint", "learned"}
	actionLinePattern = regexp.MustCompile(`^(` + strings.Join(ActionTypes, "|") + `)\(([^)]+)\): (.+)$`)
)

type ActionLine struct {
	Type  string
	Scope string
	Text  string
}

func ParseActionLine(line string) (ActionLine, bool) {
	match := actionLinePattern.FindStringSubmatch(line)
	if match == nil {
		return ActionLine{}, false
	}

	return ActionLine{Type: match[1], Scope: match[2], Text: match[3]}, true
}

type Violation struct {
	Line    int
	Code    string
	Message string
}

func Validate(value string) []Violation {
	parsed := parse(value)

	violations := []Violation{}

	if parsed.subject == "" {
		return []Violation{{
			Line:    1,
			Code:    "invalid-subject",
			Message: "subject is required",
		}}
	}

	if utf8.RuneCountInString(parsed.subject) > 72 {
		violations = append(violations, Violation{
			Line:    1,
			Code:    "subject-too-long",
			Message: fmt.Sprintf("subject must be 72 characters or fewer, got %d", utf8.RuneCountInString(parsed.subject)),
		})
	}

	if !subjectPattern.MatchString(parsed.subject) {
		violations = append(violations, Violation{
			Line:    1,
			Code:    "invalid-subject",
			Message: "subject must match <type>(<scope>): <description> with type one of " + strings.Join(subjectTypes, ", "),
		})
	}

	for _, line := range parsed.bodyLines {
		if actionLinePattern.MatchString(line.value) {
			continue
		}

		violations = append(violations, Violation{
			Line:    line.number,
			Code:    "invalid-action-format",
			Message: "body lines must match <action-type>(<scope>): <description> with action-type one of " + strings.Join(ActionTypes, ", "),
		})
	}

	return violations
}
