package lint

import (
	"strings"
	"unicode"
)

func fix(value string) string {
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

	parsed := parse(strings.Join(lines, "\n"))
	parsed.subject = strings.TrimSuffix(parsed.subject, ".")

	sections := []string{parsed.subject}
	if len(parsed.bodyLines) > 0 {
		bodyLines := make([]string, 0, len(parsed.bodyLines))
		for _, line := range parsed.bodyLines {
			bodyLines = append(bodyLines, line.value)
		}
		sections = append(sections, strings.Join(bodyLines, "\n"))
	}
	if len(parsed.trailers) > 0 {
		sections = append(sections, strings.Join(parsed.trailers, "\n"))
	}

	return strings.Join(sections, "\n\n")
}
