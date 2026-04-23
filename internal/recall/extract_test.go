package recall

import (
	"testing"
)

func TestExtractActionLines(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		wantLen   int
		wantType  string
		wantScope string
		wantDesc  string
	}{
		{
			name:      "intent line",
			lines:     []string{"intent(auth): add social login"},
			wantLen:   1,
			wantType:  "intent",
			wantScope: "auth",
			wantDesc:  "add social login",
		},
		{
			name:      "decision line",
			lines:     []string{"decision(auth): use passport.js over auth0-sdk"},
			wantLen:   1,
			wantType:  "decision",
			wantScope: "auth",
			wantDesc:  "use passport.js over auth0-sdk",
		},
		{
			name:      "rejected line",
			lines:     []string{"rejected(auth): auth0-sdk — session model incompatible"},
			wantLen:   1,
			wantType:  "rejected",
			wantScope: "auth",
			wantDesc:  "auth0-sdk — session model incompatible",
		},
		{
			name:      "constraint line",
			lines:     []string{"constraint(redis): TTL 24h max"},
			wantLen:   1,
			wantType:  "constraint",
			wantScope: "redis",
			wantDesc:  "TTL 24h max",
		},
		{
			name:      "learned line",
			lines:     []string{"learned(stripe): presentment ≠ settlement currency"},
			wantLen:   1,
			wantType:  "learned",
			wantScope: "stripe",
			wantDesc:  "presentment ≠ settlement currency",
		},
		{
			name:    "multiple lines",
			lines:   []string{"intent(auth): social login", "rejected(auth): auth0-sdk — bad fit"},
			wantLen: 2,
		},
		{
			name:    "non-action line is skipped",
			lines:   []string{"this is not an action line", "intent(x): valid"},
			wantLen: 1,
		},
		{
			name:    "empty input",
			lines:   nil,
			wantLen: 0,
		},
		{
			name:    "subject line format not matched",
			lines:   []string{"feat(auth): add login"},
			wantLen: 0,
		},
		{
			name:      "hyphenated scope",
			lines:     []string{"intent(auth-tokens): refresh token flow"},
			wantLen:   1,
			wantScope: "auth-tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractActionLines(tt.lines, "abc123", "feat(x): test")
			if len(got) != tt.wantLen {
				t.Fatalf("got %d action lines, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen == 0 {
				return
			}
			if tt.wantType != "" && got[0].Type != tt.wantType {
				t.Errorf("type = %q, want %q", got[0].Type, tt.wantType)
			}
			if tt.wantScope != "" && got[0].Scope != tt.wantScope {
				t.Errorf("scope = %q, want %q", got[0].Scope, tt.wantScope)
			}
			if tt.wantDesc != "" && got[0].Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", got[0].Description, tt.wantDesc)
			}
			// Verify commit provenance is set
			if got[0].CommitSHA != "abc123" {
				t.Errorf("commitSHA = %q, want %q", got[0].CommitSHA, "abc123")
			}
		})
	}
}

func TestParseCommitLog(t *testing.T) {
	input := "abc123\nfeat(auth): add login\nintent(auth): social login\n---COMMIT_END---\ndef456\nfix(auth): fix token\n---COMMIT_END---"
	entries := parseCommitLog(input)
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].SHA != "abc123" {
		t.Errorf("first SHA = %q, want abc123", entries[0].SHA)
	}
	if entries[0].Subject != "feat(auth): add login" {
		t.Errorf("first subject = %q", entries[0].Subject)
	}
	if len(entries[0].ActionLines) != 1 {
		t.Fatalf("first entry has %d action lines, want 1", len(entries[0].ActionLines))
	}
	if entries[0].ActionLines[0].Type != "intent" {
		t.Errorf("action type = %q, want intent", entries[0].ActionLines[0].Type)
	}
	// Second commit has no action lines
	if len(entries[1].ActionLines) != 0 {
		t.Errorf("second entry has %d action lines, want 0", len(entries[1].ActionLines))
	}
}

func TestParseCommitLogEmpty(t *testing.T) {
	entries := parseCommitLog("")
	if entries != nil {
		t.Errorf("expected nil for empty input, got %v", entries)
	}
}
