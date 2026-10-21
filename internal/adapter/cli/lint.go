package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/h3y6e/cxg/internal/core/lint"
	"github.com/h3y6e/cxg/internal/core/message"
	"github.com/spf13/cobra"
)

type linter interface {
	Lint(lint.Input) (lint.Result, error)
}

type lintOptions struct {
	messages []string
	trailers []string
	fix      bool
}

func newLintCmd(rootOpts *rootOptions, linter linter) *cobra.Command {
	opts := lintOptions{}

	cmd := &cobra.Command{
		Use:   "lint [file]",
		Short: "Validate a contextual commit message",
		Long: `Validate a contextual commit message.

The message comes from --message lines, the file argument, or stdin, in that order.
A valid message is printed to stdout; violations go to stderr, or to stdout with --json, with exit status 1.`,
		Example: `  cxg lint -m 'feat(auth): add login' -m 'intent(auth): support social login' | git commit -F -`,
		Args:    usageArgs(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(cmd, args, *rootOpts, opts, linter)
		},
	}

	cmd.Flags().StringArrayVarP(&opts.messages, "message", "m", nil, "Commit message line (repeatable)")
	cmd.Flags().StringArrayVar(&opts.trailers, "trailer", nil, "Trailer line to append (repeatable)")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Fix common formatting issues before linting")

	return cmd
}

func runLint(cmd *cobra.Command, args []string, rootOpts rootOptions, opts lintOptions, linter linter) error {
	input := lint.Input{
		Messages: opts.messages,
		Trailers: opts.trailers,
		Fix:      opts.fix,
	}
	if len(args) > 0 {
		input.FilePath = args[0]
	} else if stdin := cmd.InOrStdin(); hasReadableStdin(stdin) {
		input.Stdin = stdin
	}

	result, err := linter.Lint(input)
	if err != nil {
		if errors.Is(err, lint.ErrNoInput) {
			return usageError(err)
		}

		return err
	}

	switch {
	case rootOpts.json:
		err = writeLintJSON(cmd.OutOrStdout(), result)
	case len(result.Violations) > 0:
		err = writeViolations(cmd.ErrOrStderr(), result.Violations)
	default:
		_, err = io.WriteString(cmd.OutOrStdout(), result.Message)
	}

	if err != nil {
		return err
	}

	if len(result.Violations) > 0 {
		return exitError{status: 1}
	}

	return nil
}

func hasReadableStdin(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return true
	}

	stat, err := file.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice == 0
}

type lintResponse struct {
	Valid   bool                    `json:"valid"`
	Message string                  `json:"message,omitempty"`
	Errors  []lintViolationResponse `json:"errors"`
	Error   *failureResponse        `json:"error,omitempty"`
}

type lintViolationResponse struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeLintJSON(w io.Writer, result lint.Result) error {
	violations := make([]lintViolationResponse, 0, len(result.Violations))
	for _, violation := range result.Violations {
		violations = append(violations, lintViolationResponse{
			Line:    violation.Line,
			Code:    violation.Code,
			Message: violation.Message,
		})
	}

	response := lintResponse{
		Valid:  len(result.Violations) == 0,
		Errors: violations,
	}
	if response.Valid {
		response.Message = result.Message
	}

	return json.NewEncoder(w).Encode(response)
}

func writeViolations(w io.Writer, violations []message.Violation) error {
	for _, violation := range violations {
		if _, err := fmt.Fprintf(w, "line %d [%s] %s\n", violation.Line, violation.Code, violation.Message); err != nil {
			return err
		}
	}

	return nil
}
