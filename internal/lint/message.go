package lint

import (
	"strings"
	"unicode"
)

type message struct {
	subject   string
	bodyLines []messageLine
	trailers  []string
}

type messageLine struct {
	number int
	value  string
}

func parse(value string) message {
	result := message{
		bodyLines: []messageLine{},
		trailers:  []string{},
	}
	normalized := trimFinalNewlines(strings.ReplaceAll(value, "\r\n", "\n"))
	if normalized == "" {
		return result
	}

	lines := strings.Split(normalized, "\n")
	result.subject = lines[0]
	if len(lines) == 1 {
		return result
	}

	rest := make([]messageLine, 0, len(lines)-1)
	for index, line := range lines[1:] {
		rest = append(rest, messageLine{
			number: index + 2,
			value:  line,
		})
	}
	rest = trimTrailingEmptyLines(rest)

	if trailerStart := findTrailerStart(rest); trailerStart >= 0 {
		result.trailers = collectNonEmptyValues(rest[trailerStart:])
		rest = trimTrailingEmptyLines(rest[:trailerStart])
	}

	result.bodyLines = collectNonEmptyLines(rest)
	return result
}

func findTrailerStart(lines []messageLine) int {
	if len(lines) == 0 || !isTrailerLine(lines[len(lines)-1].value) {
		return -1
	}

	start := len(lines) - 1
	for start >= 0 && isTrailerLine(lines[start].value) {
		start--
	}
	if start < 0 || lines[start].value != "" {
		return -1
	}

	return start + 1
}

func collectNonEmptyLines(lines []messageLine) []messageLine {
	result := make([]messageLine, 0, len(lines))
	for _, line := range lines {
		if line.value != "" {
			result = append(result, line)
		}
	}

	return result
}

func collectNonEmptyValues(lines []messageLine) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.value != "" {
			result = append(result, line.value)
		}
	}

	return result
}

func trimTrailingEmptyLines(lines []messageLine) []messageLine {
	end := len(lines)
	for end > 0 && lines[end-1].value == "" {
		end--
	}

	return lines[:end]
}

func isTrailerLine(line string) bool {
	token, value, ok := strings.Cut(line, ":")
	if !ok || token == "" || value == "" {
		return false
	}

	if strings.ContainsAny(token, "()") {
		return false
	}
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			continue
		}
		return false
	}

	return value[0] == ' '
}
