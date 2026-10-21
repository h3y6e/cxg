package cli

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/core/lint"
)

func TestRecallCommand(t *testing.T) {
	t.Parallel()

	const log = `commit 1111111111111111111111111111111111111111
Author: Alice <alice@example.com>
Date:   Mon Oct 20 10:00:00 2026 +0900

    feat(auth): add login

    intent(auth): support social login
    rejected(auth): auth0-sdk session model

commit 2222222222222222222222222222222222222222
Author: Alice <alice@example.com>
Date:   Sun Oct 19 10:00:00 2026 +0900

    fix(payments): round amounts

    constraint(payments): integer cents only
`

	tests := []struct {
		name       string
		args       []string
		stdin      io.Reader
		wantStatus int
		wantStdout string
		wantFix    string
	}{
		{
			name:       "when git log is piped, executing recall prints each commit with its action lines",
			args:       []string{"recall"},
			stdin:      strings.NewReader(log),
			wantStdout: "1111111 feat(auth): add login\n  intent(auth): support social login\n  rejected(auth): auth0-sdk session model\n\n2222222 fix(payments): round amounts\n  constraint(payments): integer cents only\n",
		},
		{
			name:       "when json is requested with a query, executing recall writes the matching commits",
			args:       []string{"recall", "--json", "rejected(auth)"},
			stdin:      strings.NewReader(log),
			wantStdout: `{"schema":1,"commits":[{"hash":"1111111111111111111111111111111111111111","subject":"feat(auth): add login","actions":[{"type":"rejected","scope":"auth","text":"auth0-sdk session model"}]}]}` + "\n",
		},
		{
			name:       "when nothing matches, executing recall prints nothing and succeeds",
			args:       []string{"recall", "billing"},
			stdin:      strings.NewReader(log),
			wantStdout: "",
		},
		{
			name:       "when stdin is not plain git log output, executing recall fails with a git log fix",
			args:       []string{"recall"},
			stdin:      strings.NewReader("1111111 feat(auth): add login\n"),
			wantStatus: 2,
			wantFix:    "git log | cxg recall",
		},
		{
			name:       "when stdin is empty and not a terminal, executing recall writes an empty json record",
			args:       []string{"recall", "--json", "auth"},
			stdin:      strings.NewReader(""),
			wantStdout: `{"schema":1,"commits":[]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			stdout, _, err := executeRoot(t, lint.New(os.ReadFile), tt.stdin, tt.args...)

			// Assert
			if status := exitStatus(err); status != tt.wantStatus {
				t.Fatalf("exit status = %d (error %v), want %d", status, err, tt.wantStatus)
			}

			if stdout != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout, tt.wantStdout)
			}

			if exitErr, _ := errors.AsType[exitError](err); exitErr.fix != tt.wantFix {
				t.Fatalf("fix = %q, want %q", exitErr.fix, tt.wantFix)
			}
		})
	}
}
