package recall

import (
	"fmt"
	"sort"
	"strings"
)

// ActionTypePriority defines display order for action types.
// "Don't do this" signals first, then "do this" signals.
var ActionTypePriority = []string{"rejected", "constraint", "intent", "decision", "learned"}

// Format renders a RecallResult as plain text.
func Format(result RecallResult) string {
	switch result.Mode {
	case ModeScope:
		return formatScopeQuery(result)
	case ModeActionScope:
		return formatActionScopeQuery(result)
	default:
		return formatDefault(result)
	}
}

func formatDefault(r RecallResult) string {
	var b strings.Builder

	if len(r.ActionLines) > 0 {
		writeGroupedByType(&b, r.ActionLines)
	} else {
		b.WriteString("No action lines found.\n")
	}

	return b.String()
}

func formatScopeQuery(r RecallResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Scope: %s", r.Query)
	if len(r.Scopes) > 1 {
		others := make([]string, 0, len(r.Scopes))
		for _, s := range r.Scopes {
			if s != r.Query {
				others = append(others, s)
			}
		}
		if len(others) > 0 {
			sort.Strings(others)
			fmt.Fprintf(&b, " (also found: %s)", strings.Join(others, ", "))
		}
	}
	b.WriteString("\n")

	if len(r.ActionLines) == 0 {
		b.WriteString("No action lines found.\n")
		return b.String()
	}

	b.WriteString("\n")
	writeGroupedByType(&b, r.ActionLines)
	return b.String()
}

func formatActionScopeQuery(r RecallResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s across history:\n", r.Query)

	if len(r.ActionLines) == 0 {
		b.WriteString("No matches found.\n")
		return b.String()
	}

	b.WriteString("\n")
	for _, al := range r.ActionLines {
		fmt.Fprintf(&b, "  - %s: %s\n", al.Scope, al.Description)
		fmt.Fprintf(&b, "    from: %s\n", al.CommitSubject)
	}
	return b.String()
}

func writeGroupedByType(b *strings.Builder, actions []ActionLine) {
	grouped := make(map[string][]ActionLine)
	for _, al := range actions {
		grouped[al.Type] = append(grouped[al.Type], al)
	}

	for _, typ := range ActionTypePriority {
		lines, ok := grouped[typ]
		if !ok {
			continue
		}
		fmt.Fprintf(b, "%s:\n", typ)
		for _, al := range lines {
			fmt.Fprintf(b, "  - %s: %s\n", al.Scope, al.Description)
		}
	}
}
