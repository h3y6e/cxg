package lint

import (
	"fmt"
	"regexp"

	"github.com/h3y6e/cxg/internal/message"
)

var (
	subjectPattern    = regexp.MustCompile(`^(feat|fix|refactor|perf|test|docs|style|build|ci|chore|revert)(\([^)]+\))?(!)?: .+$`)
	actionLinePattern = regexp.MustCompile(`^(intent|decision|rejected|constraint|learned)\([^)]+\): .+$`)
)

func Validate(value string) []message.ValidationError {
	parsed, _ := message.Parse(value)

	var errors []message.ValidationError
	if parsed.Subject == "" {
		return []message.ValidationError{{
			Line:    1,
			Code:    "invalid-subject",
			Message: "subject is required",
		}}
	}

	if len(parsed.Subject) > 72 {
		errors = append(errors, message.ValidationError{
			Line:    1,
			Code:    "subject-too-long",
			Message: fmt.Sprintf("subject must be 72 characters or fewer, got %d", len(parsed.Subject)),
		})
	}

	if !subjectPattern.MatchString(parsed.Subject) {
		errors = append(errors, message.ValidationError{
			Line:    1,
			Code:    "invalid-subject",
			Message: "subject must match <type>(<scope>): <description>",
		})
	}

	for index, line := range parsed.BodyLines {
		if actionLinePattern.MatchString(line) {
			continue
		}

		errors = append(errors, message.ValidationError{
			Line:    3 + index,
			Code:    "invalid-action-format",
			Message: "body lines must match <action-type>(<scope>): <description>",
		})
	}

	return errors
}
