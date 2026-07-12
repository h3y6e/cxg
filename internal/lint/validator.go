package lint

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	subjectPattern    = regexp.MustCompile(`^(feat|fix|refactor|perf|test|docs|style|build|ci|chore|revert)(\([^)]+\))?(!)?: .+$`)
	actionLinePattern = regexp.MustCompile(`^(intent|decision|rejected|constraint|learned)\([^)]+\): .+$`)
)

func validate(value string) []Violation {
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
			Message: "subject must match <type>(<scope>): <description>",
		})
	}

	for _, line := range parsed.bodyLines {
		if actionLinePattern.MatchString(line.value) {
			continue
		}

		violations = append(violations, Violation{
			Line:    line.number,
			Code:    "invalid-action-format",
			Message: "body lines must match <action-type>(<scope>): <description>",
		})
	}

	return violations
}
