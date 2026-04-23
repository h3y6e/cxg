package recall

import (
	"strings"
	"testing"
)

func TestParseArgument_Default(t *testing.T) {
	req := ParseArgument("")
	if req.Mode != ModeDefault {
		t.Errorf("mode = %d, want ModeDefault", req.Mode)
	}
}

func TestParseArgument_Scope(t *testing.T) {
	req := ParseArgument("auth")
	if req.Mode != ModeScope {
		t.Errorf("mode = %d, want ModeScope", req.Mode)
	}
	if req.Scope != "auth" {
		t.Errorf("scope = %q, want %q", req.Scope, "auth")
	}
}

func TestParseArgument_ActionScope(t *testing.T) {
	req := ParseArgument("rejected(auth)")
	if req.Mode != ModeActionScope {
		t.Errorf("mode = %d, want ModeActionScope", req.Mode)
	}
	if req.ActionType != "rejected" {
		t.Errorf("actionType = %q, want %q", req.ActionType, "rejected")
	}
	if req.Scope != "auth" {
		t.Errorf("scope = %q, want %q", req.Scope, "auth")
	}
	if req.Query != "rejected(auth)" {
		t.Errorf("query = %q, want %q", req.Query, "rejected(auth)")
	}
}

func TestParseArgument_ActionScopeHyphen(t *testing.T) {
	req := ParseArgument("constraint(auth-tokens)")
	if req.Mode != ModeActionScope {
		t.Errorf("mode = %d, want ModeActionScope", req.Mode)
	}
	if req.ActionType != "constraint" {
		t.Errorf("actionType = %q, want %q", req.ActionType, "constraint")
	}
	if req.Scope != "auth-tokens" {
		t.Errorf("scope = %q, want %q", req.Scope, "auth-tokens")
	}
}

func TestFormatDefault_FeatureBranch(t *testing.T) {
	result := RecallResult{
		Mode: ModeDefault,
		Branch: BranchState{
			Current:      "feat/auth",
			Base:         "main",
			CommitsAhead: 3,
		},
		ActionLines: []ActionLine{
			{Type: "intent", Scope: "auth", Description: "add social login"},
			{Type: "rejected", Scope: "auth", Description: "auth0-sdk — bad fit"},
		},
	}
	output := Format(result)
	// rejected before intent (priority order)
	if !strings.Contains(output, "rejected:\n  - auth: auth0-sdk") {
		t.Errorf("missing rejected with scope in:\n%s", output)
	}
	if !strings.Contains(output, "intent:\n  - auth: add social login") {
		t.Errorf("missing intent with scope in:\n%s", output)
	}
}

func TestFormatDefault_FeatureBranchEmpty(t *testing.T) {
	result := RecallResult{
		Mode: ModeDefault,
		Branch: BranchState{
			Current:      "feat/new",
			Base:         "main",
			CommitsAhead: 0,
			HasUnstaged:  true,
		},
		ActionLines: []ActionLine{
			{Type: "intent", Scope: "auth", Description: "social login from main"},
			{Type: "constraint", Scope: "auth", Description: "redis TTL 24h"},
			{Type: "learned", Scope: "stripe", Description: "presentment != settlement"},
		},
	}
	output := Format(result)
	if !strings.Contains(output, "constraint:\n  - auth: redis TTL 24h") {
		t.Errorf("missing constraint with scope in:\n%s", output)
	}
	if !strings.Contains(output, "intent:\n  - auth: social login from main") {
		t.Errorf("missing intent with scope in:\n%s", output)
	}
	if !strings.Contains(output, "learned:\n  - stripe: presentment != settlement") {
		t.Errorf("missing learned with scope in:\n%s", output)
	}
}

func TestFormatDefault_NoActionLines(t *testing.T) {
	result := RecallResult{
		Mode: ModeDefault,
		Branch: BranchState{
			Current:      "feat/new",
			Base:         "main",
			CommitsAhead: 2,
		},
		Commits: []CommitEntry{
			{SHA: "a1", Subject: "feat(auth): add login"},
			{SHA: "a2", Subject: "fix(auth): fix token"},
		},
	}
	output := Format(result)
	if !strings.Contains(output, "No action lines found.") {
		t.Errorf("missing fallback message in:\n%s", output)
	}
}

func TestFormatDefault_DefaultBranch(t *testing.T) {
	result := RecallResult{
		Mode: ModeDefault,
		Branch: BranchState{
			Current:   "main",
			Base:      "main",
			IsDefault: true,
		},
		ActionLines: []ActionLine{
			{Type: "intent", Scope: "auth", Description: "social login"},
			{Type: "learned", Scope: "stripe", Description: "presentment != settlement"},
		},
	}
	output := Format(result)
	if !strings.Contains(output, "intent:\n  - auth: social login") {
		t.Errorf("missing intent with scope in:\n%s", output)
	}
	if !strings.Contains(output, "learned:\n  - stripe: presentment != settlement") {
		t.Errorf("missing learned with scope in:\n%s", output)
	}
}

func TestFormatScopeQuery(t *testing.T) {
	result := RecallResult{
		Mode:  ModeScope,
		Query: "auth",
		ActionLines: []ActionLine{
			{Type: "intent", Scope: "auth", Description: "social login"},
			{Type: "rejected", Scope: "auth-tokens", Description: "JWT too short"},
		},
		Scopes: []string{"auth", "auth-tokens"},
	}
	output := Format(result)
	if !strings.Contains(output, "Scope: auth (also found: auth-tokens)") {
		t.Errorf("missing scope header in:\n%s", output)
	}
	if !strings.Contains(output, "rejected:\n  - auth-tokens: JWT too short") {
		t.Errorf("missing rejected with scope in:\n%s", output)
	}
	if !strings.Contains(output, "intent:\n  - auth: social login") {
		t.Errorf("missing intent with scope in:\n%s", output)
	}
}

func TestFormatScopeQuery_NoMatches(t *testing.T) {
	result := RecallResult{
		Mode:  ModeScope,
		Query: "payments",
	}
	output := Format(result)
	if !strings.Contains(output, "No action lines found") {
		t.Errorf("missing no-match message in:\n%s", output)
	}
}

func TestFormatActionScopeQuery(t *testing.T) {
	result := RecallResult{
		Mode:  ModeActionScope,
		Query: "rejected(auth)",
		ActionLines: []ActionLine{
			{Type: "rejected", Scope: "auth", Description: "auth0-sdk — bad fit", CommitSubject: "feat(auth): add login"},
		},
	}
	output := Format(result)
	if !strings.Contains(output, "rejected(auth) across history:") {
		t.Errorf("missing header in:\n%s", output)
	}
	if !strings.Contains(output, "  - auth: auth0-sdk") {
		t.Errorf("missing description with scope in:\n%s", output)
	}
	if !strings.Contains(output, "from: feat(auth): add login") {
		t.Errorf("missing provenance in:\n%s", output)
	}
}

func TestFormatActionScopeQuery_NoMatches(t *testing.T) {
	result := RecallResult{
		Mode:  ModeActionScope,
		Query: "rejected(unknown)",
	}
	output := Format(result)
	if !strings.Contains(output, "No matches found") {
		t.Errorf("missing no-match message in:\n%s", output)
	}
}
