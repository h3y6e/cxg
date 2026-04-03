package message

import (
	"strings"
	"unicode"
)

func Fix(value string) string {
	normalized := trimFinalNewlines(strings.ReplaceAll(value, "\r\n", "\n"))
	if normalized == "" {
		return ""
	}

	lines := strings.Split(normalized, "\n")
	for index, line := range lines {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if index > 0 && line != "" {
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
		}
		lines[index] = line
	}

	parsed, _ := Parse(strings.Join(lines, "\n"))
	parsed.Subject = strings.TrimSuffix(parsed.Subject, ".")

	sections := []string{parsed.Subject}
	if len(parsed.BodyLines) > 0 {
		sections = append(sections, strings.Join(parsed.BodyLines, "\n"))
	}
	if len(parsed.Trailers) > 0 {
		sections = append(sections, strings.Join(parsed.Trailers, "\n"))
	}

	return strings.Join(sections, "\n\n")
}
