package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/h3y6e/cxg/internal/lint"
	"github.com/h3y6e/cxg/internal/message"
	"github.com/spf13/cobra"
)

type lintOptions struct {
	messages []string
	trailers []string
	fix      bool
}

type ExitError struct {
	Code int
}

func (err ExitError) Error() string {
	return fmt.Sprintf("exit %d", err.Code)
}

func newLintCmd(rootOpts *rootOptions) *cobra.Command {
	opts := lintOptions{}

	cmd := &cobra.Command{
		Use:   "lint [file]",
		Short: "Validate a contextual commit message",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(cmd, args, *rootOpts, opts)
		},
	}

	cmd.Flags().StringArrayVarP(&opts.messages, "message", "m", nil, "Commit message line (repeatable)")
	cmd.Flags().StringArrayVar(&opts.trailers, "trailer", nil, "Trailer line to append (repeatable)")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Fix common formatting issues before linting")

	return cmd
}

func runLint(cmd *cobra.Command, args []string, rootOpts rootOptions, opts lintOptions) error {
	input := message.Input{
		Messages: opts.messages,
		Stdin:    cmd.InOrStdin(),
		HasStdin: hasReadableStdin(cmd),
		Trailers: opts.trailers,
	}
	if len(args) > 0 {
		input.FilePath = args[0]
	}

	msg, err := message.Resolve(input)
	if err != nil {
		return err
	}

	if opts.fix {
		msg = message.Fix(msg)
	}

	validationErrors := lint.Validate(msg)
	if len(validationErrors) > 0 {
		if rootOpts.json {
			return writeLintJSON(cmd, lintResponse{
				Valid:  false,
				Errors: validationErrors,
			}, true)
		}

		if err := writeValidationErrors(cmd, validationErrors); err != nil {
			return err
		}

		return ExitError{Code: 1}
	}

	if rootOpts.json {
		return writeLintJSON(cmd, lintResponse{
			Valid:   true,
			Message: msg,
			Errors:  []message.ValidationError{},
		}, false)
	}

	_, err = fmt.Fprint(cmd.OutOrStdout(), msg)
	return err
}

func hasReadableStdin(cmd *cobra.Command) bool {
	if cmd.InOrStdin() != os.Stdin {
		return true
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice == 0
}

type lintResponse struct {
	Valid   bool                      `json:"valid"`
	Message string                    `json:"message,omitempty"`
	Errors  []message.ValidationError `json:"errors"`
}

func writeLintJSON(cmd *cobra.Command, response lintResponse, exitWithError bool) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	if err := encoder.Encode(response); err != nil {
		return err
	}
	if exitWithError {
		return ExitError{Code: 1}
	}

	return nil
}

func writeValidationErrors(cmd *cobra.Command, validationErrors []message.ValidationError) error {
	for _, validationError := range validationErrors {
		_, err := fmt.Fprintf(
			cmd.ErrOrStderr(),
			"line %d [%s] %s\n",
			validationError.Line,
			validationError.Code,
			validationError.Message,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
