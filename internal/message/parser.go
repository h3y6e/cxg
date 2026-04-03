package message

import (
	"strings"
	"unicode"
)

func Parse(value string) (CommitMessage, error) {
	normalized := trimFinalNewlines(strings.ReplaceAll(value, "\r\n", "\n"))
	if normalized == "" {
		return CommitMessage{}, nil
	}

	lines := strings.Split(normalized, "\n")
	result := CommitMessage{
		Subject: lines[0],
	}

	if len(lines) == 1 {
		return result, nil
	}

	rest := lines[1:]
	rest = trimTrailingEmptyLines(rest)

	trailerStart := findTrailerStart(rest)
	if trailerStart >= 0 {
		result.Trailers = collectNonEmpty(rest[trailerStart:])
		rest = trimTrailingEmptyLines(rest[:trailerStart])
	}

	result.BodyLines = collectNonEmpty(rest)

	return result, nil
}

func findTrailerStart(lines []string) int {
	if len(lines) == 0 {
		return -1
	}

	last := len(lines) - 1
	if !isTrailerLine(lines[last]) {
		return -1
	}

	start := last
	for start >= 0 && isTrailerLine(lines[start]) {
		start--
	}

	if start < 0 || lines[start] != "" {
		return -1
	}

	return start + 1
}

func collectNonEmpty(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		result = append(result, line)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func trimTrailingEmptyLines(lines []string) []string {
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}

	return lines[:end]
}

func isTrailerLine(line string) bool {
	colon := strings.IndexByte(line, ':')
	if colon <= 0 || colon == len(line)-1 {
		return false
	}

	token := line[:colon]
	if strings.ContainsAny(token, "()") {
		return false
	}

	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			continue
		}
		return false
	}

	return line[colon+1] == ' '
}
