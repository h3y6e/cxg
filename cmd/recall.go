package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	internalgit "github.com/h3y6e/cxg/internal/git"
	"github.com/h3y6e/cxg/internal/recall"
	"github.com/spf13/cobra"
)

func newRecallCmd(rootOpts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recall [scope | action(scope)]",
		Short: "Extract contextual action lines from commit history",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecall(cmd, args, *rootOpts)
		},
	}

	return cmd
}

type recallResponse struct {
	Mode        string                 `json:"mode"`
	Branch      *recallBranchResponse  `json:"branch,omitempty"`
	Query       string                 `json:"query,omitempty"`
	Scopes      []string               `json:"scopes,omitempty"`
	ActionLines []recallActionResponse `json:"actionLines"`
}

type recallBranchResponse struct {
	Current      string `json:"current"`
	Base         string `json:"base"`
	CommitsAhead int    `json:"commitsAhead"`
	IsDefault    bool   `json:"isDefault"`
	HasUnstaged  bool   `json:"hasUnstaged"`
	HasStaged    bool   `json:"hasStaged"`
}

type recallActionResponse struct {
	Type          string `json:"type"`
	Scope         string `json:"scope"`
	Description   string `json:"description"`
	CommitSHA     string `json:"commitSHA"`
	CommitSubject string `json:"commitSubject"`
}

func runRecall(cmd *cobra.Command, args []string, rootOpts rootOptions) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if !internalgit.IsRepository(cmd.Context(), workingDir) {
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "not a git repository"); err != nil {
			return err
		}
		return ExitError{Code: 1}
	}

	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	req := recall.ParseArgument(arg)
	req.Dir = workingDir

	result, err := recall.Run(cmd.Context(), req)
	if err != nil {
		if errors.Is(err, recall.ErrEmptyRepository) {
			if _, writeErr := fmt.Fprintln(cmd.ErrOrStderr(), "repository has no commits"); writeErr != nil {
				return writeErr
			}
			return ExitError{Code: 1}
		}
		return err
	}

	if rootOpts.json {
		return writeRecallJSON(cmd, result)
	}

	output := recall.Format(result)
	_, err = fmt.Fprint(cmd.OutOrStdout(), output)
	return err
}

func writeRecallJSON(cmd *cobra.Command, result recall.RecallResult) error {
	response := recallResponse{
		Mode:        modeString(result.Mode),
		Query:       result.Query,
		Scopes:      result.Scopes,
		ActionLines: make([]recallActionResponse, 0, len(result.ActionLines)),
	}

	if result.Mode == recall.ModeDefault {
		response.Branch = &recallBranchResponse{
			Current:      result.Branch.Current,
			Base:         result.Branch.Base,
			CommitsAhead: result.Branch.CommitsAhead,
			IsDefault:    result.Branch.IsDefault,
			HasUnstaged:  result.Branch.HasUnstaged,
			HasStaged:    result.Branch.HasStaged,
		}
	}

	for _, al := range result.ActionLines {
		response.ActionLines = append(response.ActionLines, recallActionResponse{
			Type:          al.Type,
			Scope:         al.Scope,
			Description:   al.Description,
			CommitSHA:     al.CommitSHA,
			CommitSubject: al.CommitSubject,
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	return encoder.Encode(response)
}

func modeString(m recall.Mode) string {
	switch m {
	case recall.ModeScope:
		return "scope"
	case recall.ModeActionScope:
		return "action-scope"
	default:
		return "default"
	}
}
