package recall_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/h3y6e/cxg/internal/core/message"
	"github.com/h3y6e/cxg/internal/core/recall"
)

const gitLog = `commit 1111111111111111111111111111111111111111 (HEAD -> main)
Merge: aaaaaaa bbbbbbb
Author: Alice <alice@example.com>
Date:   Mon Oct 20 10:00:00 2026 +0900

    feat(auth): add login

    intent(auth): support social login
    rejected(a+b): literal plus scope
    Co-authored-by: Bob <bob@example.com>

 auth.go | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)

commit 2222222222222222222222222222222222222222
Author: Alice <alice@example.com>
Date:   Sun Oct 19 10:00:00 2026 +0900

    fix(ab): plain scope

    decision(ab): keep it
    rejected(auth-tokens): opaque tokens

diff --git a/ab.go b/ab.go
--- a/ab.go
+++ b/ab.go
@@ -1 +1 @@
-    intent(auth): inside a diff
+    intent(auth): still inside a diff

commit 3333333333333333333333333333333333333333
Author: Alice <alice@example.com>
Date:   Sat Oct 18 10:00:00 2026 +0900

    chore: no context
`

func TestRecall(t *testing.T) {
	t.Parallel()

	login := func(actions ...message.ActionLine) recall.Commit {
		return recall.Commit{Hash: strings.Repeat("1", 40), Subject: "feat(auth): add login", Actions: actions}
	}
	plain := func(actions ...message.ActionLine) recall.Commit {
		return recall.Commit{Hash: strings.Repeat("2", 40), Subject: "fix(ab): plain scope", Actions: actions}
	}

	tests := []struct {
		name  string
		input string
		query recall.Query
		want  []recall.Commit
	}{
		{
			name:  "when no query is given, recalling returns every action line from the message blocks in input order",
			input: gitLog,
			want: []recall.Commit{
				login(
					message.ActionLine{Type: "intent", Scope: "auth", Text: "support social login"},
					message.ActionLine{Type: "rejected", Scope: "a+b", Text: "literal plus scope"},
				),
				plain(
					message.ActionLine{Type: "decision", Scope: "ab", Text: "keep it"},
					message.ActionLine{Type: "rejected", Scope: "auth-tokens", Text: "opaque tokens"},
				),
			},
		},
		{
			name:  "when a scope is given, recalling keeps action lines whose scope starts with it",
			input: gitLog,
			query: recall.Query{Scope: "auth"},
			want: []recall.Commit{
				login(message.ActionLine{Type: "intent", Scope: "auth", Text: "support social login"}),
				plain(message.ActionLine{Type: "rejected", Scope: "auth-tokens", Text: "opaque tokens"}),
			},
		},
		{
			name:  "when a scope has regexp metacharacters, recalling matches it literally",
			input: gitLog,
			query: recall.Query{Scope: "a+b"},
			want: []recall.Commit{
				login(message.ActionLine{Type: "rejected", Scope: "a+b", Text: "literal plus scope"}),
			},
		},
		{
			name:  "when an action and scope are given, recalling keeps only that action type",
			input: gitLog,
			query: recall.Query{Type: "rejected", Scope: "auth"},
			want: []recall.Commit{
				plain(message.ActionLine{Type: "rejected", Scope: "auth-tokens", Text: "opaque tokens"}),
			},
		},
		{
			name:  "when the input is empty, recalling returns no commits",
			input: "",
			want:  []recall.Commit{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got, err := recall.Recall(strings.NewReader(tt.input), tt.query)

			// Assert
			if err != nil {
				t.Fatalf("Recall() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Recall() = %#v, want %#v", got, tt.want)
			}
		})
	}

	for name, input := range map[string]string{
		"oneline git log output":     "1111111 feat(auth): add login\n",
		"a commit line with no hash": "commit \n",
	} {
		t.Run("when the input is "+name+", recalling fails as not a git log", func(t *testing.T) {
			t.Parallel()

			// Act
			_, err := recall.Recall(strings.NewReader(input), recall.Query{})

			// Assert
			if !errors.Is(err, recall.ErrNotGitLog) {
				t.Fatalf("Recall() error = %v, want %v", err, recall.ErrNotGitLog)
			}
		})
	}

	t.Run("when an action line is longer than any read buffer, recalling keeps it whole", func(t *testing.T) {
		t.Parallel()

		// Arrange
		text := strings.Repeat("x", 2<<20)

		// Act
		got, err := recall.Recall(strings.NewReader("commit 1111111\n\n    feat: x\n\n    intent(a): "+text), recall.Query{})

		// Assert
		if err != nil {
			t.Fatalf("Recall() error = %v", err)
		}

		if len(got) != 1 || got[0].Actions[0].Text != text {
			t.Fatalf("Recall() lost the long action line")
		}
	})
}

func TestParseQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    recall.Query
		wantErr bool
	}{
		{name: "when the query is a bare scope, parsing returns only the scope", value: "auth", want: recall.Query{Scope: "auth"}},
		{name: "when the query is an action with a scope, parsing returns both", value: "rejected(auth)", want: recall.Query{Type: "rejected", Scope: "auth"}},
		{name: "when the action type is unknown, parsing fails", value: "nope(auth)", wantErr: true},
		{name: "when parentheses are unbalanced, parsing fails", value: "rejected(auth", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got, err := recall.ParseQuery(tt.value)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("ParseQuery() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
