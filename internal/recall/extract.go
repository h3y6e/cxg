package recall

import "regexp"

var actionLinePattern = regexp.MustCompile(
	`^(intent|decision|rejected|constraint|learned)\(([^)]+)\): (.+)$`,
)

// ExtractActionLines parses commit body lines and returns any contextual action lines found.
func ExtractActionLines(bodyLines []string, commitSHA, commitSubject string) []ActionLine {
	var result []ActionLine
	for _, line := range bodyLines {
		matches := actionLinePattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		result = append(result, ActionLine{
			Type:          matches[1],
			Scope:         matches[2],
			Description:   matches[3],
			CommitSHA:     commitSHA,
			CommitSubject: commitSubject,
		})
	}
	return result
}
