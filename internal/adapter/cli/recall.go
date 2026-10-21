package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/h3y6e/cxg/internal/core/recall"
	"github.com/spf13/cobra"
)

func newRecallCmd(rootOpts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "recall [query]",
		Short: "Extract contextual action lines from git log output on stdin",
		Long: "Extract contextual action lines from git log output on stdin.\n\n" +
			"The query is <scope> or <action>(<scope>); scopes match by literal prefix.\n" +
			"Pipe plain git log output: --oneline, --format, and --pretty are not supported.",
		Example: "  git log main..HEAD | cxg recall\n" +
			"  git log --all | cxg recall auth\n" +
			"  git log --all | cxg recall 'rejected(auth)'",
		Args: usageArgs(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecall(cmd, args, *rootOpts)
		},
	}
}

func runRecall(cmd *cobra.Command, args []string, rootOpts rootOptions) error {
	if rootOpts.terminal {
		if len(args) == 0 {
			return cmd.Help()
		}

		return exitError{status: 2, cause: errors.New("recall: no git log on stdin"), fix: "git log --all | cxg recall -- " + shellQuote(args[0])}
	}

	if err := rootOpts.checkEnv(); err != nil {
		return err
	}

	query := recall.Query{}

	if len(args) > 0 {
		parsed, err := recall.ParseQuery(args[0])
		if err != nil {
			return usageError(err)
		}

		query = parsed
	}

	commits, err := recall.Recall(cmd.InOrStdin(), query)
	if errors.Is(err, recall.ErrNotGitLog) {
		return exitError{status: 2, cause: err, fix: "git log | cxg recall"}
	}

	if err != nil {
		return err
	}

	if rootOpts.json {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(newRecallResponse(commits, nil))
	}

	return writeCommits(cmd.OutOrStdout(), commits)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

type recallResponse struct {
	Schema  int                    `json:"schema"`
	Commits []recallCommitResponse `json:"commits"`
	Error   *failureResponse       `json:"error,omitempty"`
}

type recallCommitResponse struct {
	Hash    string                 `json:"hash"`
	Subject string                 `json:"subject"`
	Actions []recallActionResponse `json:"actions"`
}

type recallActionResponse struct {
	Type  string `json:"type"`
	Scope string `json:"scope"`
	Text  string `json:"text"`
}

func newRecallResponse(commits []recall.Commit, failure *failureResponse) recallResponse {
	response := recallResponse{Schema: 1, Commits: []recallCommitResponse{}, Error: failure}

	for _, commit := range commits {
		actions := make([]recallActionResponse, 0, len(commit.Actions))
		for _, action := range commit.Actions {
			actions = append(actions, recallActionResponse{Type: action.Type, Scope: action.Scope, Text: action.Text})
		}

		response.Commits = append(response.Commits, recallCommitResponse{Hash: commit.Hash, Subject: commit.Subject, Actions: actions})
	}

	return response
}

func writeCommits(w io.Writer, commits []recall.Commit) error {
	for index, commit := range commits {
		if index > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(w, "%.7s %s\n", commit.Hash, commit.Subject); err != nil {
			return err
		}

		for _, action := range commit.Actions {
			if _, err := fmt.Fprintf(w, "  %s(%s): %s\n", action.Type, action.Scope, action.Text); err != nil {
				return err
			}
		}
	}

	return nil
}
