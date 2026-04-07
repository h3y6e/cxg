package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	internalcommit "github.com/h3y6e/cxg/internal/commit"
	"github.com/h3y6e/cxg/internal/message"
	"github.com/spf13/cobra"
)

type commitOptions struct {
	messages      []string
	trailers      []string
	all           bool
	amend         bool
	allowEmpty    bool
	signoff       bool
	author        string
	date          string
	edit          bool
	file          string
	reuseMessage  string
	reeditMessage string
	template      string
	noVerify      bool
}

func newCommitCmd(rootOpts *rootOptions) *cobra.Command {
	opts := commitOptions{}

	cmd := &cobra.Command{
		Use:   "commit [file]",
		Short: "Validate a contextual commit message and create a git commit",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(cmd, args, *rootOpts, opts)
		},
	}

	cmd.Flags().StringArrayVarP(&opts.messages, "message", "m", nil, "Commit message paragraph (repeatable)")
	cmd.Flags().StringArrayVar(&opts.trailers, "trailer", nil, "Trailer line to append (repeatable)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "Stage all tracked modified and deleted files")
	cmd.Flags().BoolVar(&opts.amend, "amend", false, "Amend the previous commit")
	cmd.Flags().BoolVar(&opts.allowEmpty, "allow-empty", false, "Allow creating an empty commit")
	cmd.Flags().BoolVar(&opts.signoff, "signoff", false, "Add Signed-off-by trailer")
	cmd.Flags().StringVar(&opts.author, "author", "", "Override the commit author")
	cmd.Flags().StringVar(&opts.date, "date", "", "Override the author date")

	cmd.Flags().BoolVarP(&opts.edit, "edit", "e", false, "Edit the commit message")
	cmd.Flags().StringVarP(&opts.file, "file", "F", "", "Take commit message from file")
	cmd.Flags().StringVarP(&opts.reuseMessage, "reuse-message", "C", "", "Reuse an existing commit message")
	cmd.Flags().StringVarP(&opts.reeditMessage, "reedit-message", "c", "", "Reuse and edit an existing commit message")
	cmd.Flags().StringVarP(&opts.template, "template", "t", "", "Use a commit message template")
	cmd.Flags().BoolVar(&opts.noVerify, "no-verify", false, "Bypass commit hooks")

	return cmd
}

func runCommit(cmd *cobra.Command, args []string, rootOpts rootOptions, opts commitOptions) error {
	if opts.edit || opts.file != "" || opts.reuseMessage != "" || opts.reeditMessage != "" || opts.template != "" || opts.noVerify {
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "unsupported git mode: commit message must be provided through cxg input flags, stdin, or file argument"); err != nil {
			return err
		}
		return ExitError{Code: 1}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	request := internalcommit.Request{
		Dir: workingDir,
		PrepareRequest: internalcommit.PrepareRequest{
			Messages: opts.messages,
			Stdin:    cmd.InOrStdin(),
			HasStdin: hasReadableStdin(cmd),
			Trailers: opts.trailers,
			Fix:      true,
		},
		All:        opts.all,
		Amend:      opts.amend,
		AllowEmpty: opts.allowEmpty,
		Signoff:    opts.signoff,
		Author:     opts.author,
		Date:       opts.date,
	}
	if len(args) > 0 {
		request.FilePath = args[0]
	}

	result, err := internalcommit.Run(cmd.Context(), request)
	if err != nil {
		return err
	}

	if !result.Valid {
		if rootOpts.json {
			return writeCommitJSON(cmd, commitResponseFromResult(result), true)
		}
		if err := writeValidationErrors(cmd, result.Errors); err != nil {
			return err
		}
		return ExitError{Code: 1}
	}

	if result.GitError != nil {
		if rootOpts.json {
			return writeCommitJSON(cmd, commitResponseFromResult(result), true)
		}
		if _, err := fmt.Fprint(cmd.OutOrStdout(), result.Stdout); err != nil {
			return err
		}
		if _, err := fmt.Fprint(cmd.ErrOrStderr(), result.Stderr); err != nil {
			return err
		}
		return ExitError{Code: result.GitError.ExitCode}
	}

	if rootOpts.json {
		return writeCommitJSON(cmd, commitResponseFromResult(result), false)
	}
	if _, err := fmt.Fprint(cmd.OutOrStdout(), result.Stdout); err != nil {
		return err
	}
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), result.Stderr); err != nil {
		return err
	}

	return nil
}

type commitResponse struct {
	Valid     bool                      `json:"valid"`
	Committed bool                      `json:"committed"`
	Message   string                    `json:"message,omitempty"`
	Errors    []message.ValidationError `json:"errors"`
	GitError  *gitErrorResponse         `json:"gitError,omitempty"`
}

type gitErrorResponse struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

func commitResponseFromResult(result internalcommit.Result) commitResponse {
	response := commitResponse{
		Valid:     result.Valid,
		Committed: result.Committed,
		Message:   result.Message,
		Errors:    result.Errors,
	}
	if result.GitError != nil {
		response.GitError = &gitErrorResponse{
			ExitCode: result.GitError.ExitCode,
			Stdout:   result.GitError.Stdout,
			Stderr:   result.GitError.Stderr,
		}
	}

	return response
}

func writeCommitJSON(cmd *cobra.Command, response commitResponse, exitWithError bool) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	if err := encoder.Encode(response); err != nil {
		return err
	}
	if exitWithError {
		code := 1
		if response.GitError != nil && response.GitError.ExitCode != 0 {
			code = response.GitError.ExitCode
		}
		return ExitError{Code: code}
	}

	return nil
}
