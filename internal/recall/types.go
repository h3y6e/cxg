package recall

type Mode int

const (
	ModeDefault     Mode = iota // No arguments — branch context briefing
	ModeScope                   // Bare word — scope query across all history
	ModeActionScope             // action(scope) — action+scope query
)

type BranchState struct {
	Current      string
	Base         string
	CommitsAhead int
	IsDefault    bool
	HasUnstaged  bool
	HasStaged    bool
}

type ActionLine struct {
	Type          string // intent, decision, rejected, constraint, learned
	Scope         string
	Description   string
	CommitSHA     string
	CommitSubject string
}

type CommitEntry struct {
	SHA         string
	Subject     string
	ActionLines []ActionLine
}

type RecallResult struct {
	Mode        Mode
	Branch      BranchState
	Query       string        // scope or action(scope) for query modes
	Commits     []CommitEntry
	ActionLines []ActionLine // flattened from Commits
	Scopes      []string     // discovered sub-scopes (scope query mode)
}
