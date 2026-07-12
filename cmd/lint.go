package cmd

import (
	"errors"
	"io"
	"os"

	"github.com/h3y6e/cxg/internal/lint"
	"github.com/spf13/cobra"
)

type lintOptions struct {
	messages []string
	trailers []string
	fix      bool
}

func newLintCmd(rootOpts *rootOptions, runner Runner) *cobra.Command {
	opts := lintOptions{}

	cmd := &cobra.Command{
		Use:   "lint [file]",
		Short: "Validate a contextual commit message",
		Args:  usageArgs(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(cmd, args, *rootOpts, opts, runner)
		},
	}

	cmd.Flags().StringArrayVarP(&opts.messages, "message", "m", nil, "Commit message line (repeatable)")
	cmd.Flags().StringArrayVar(&opts.trailers, "trailer", nil, "Trailer line to append (repeatable)")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Fix common formatting issues before linting")

	return cmd
}

func runLint(cmd *cobra.Command, args []string, rootOpts rootOptions, opts lintOptions, runner Runner) error {
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

	result, err := runner.Run(input)
	if err != nil {
		if errors.Is(err, lint.ErrNoInput) {
			return usageError(err)
		}
		return err
	}

	if rootOpts.json {
		if err := writeLintJSON(cmd, result); err != nil {
			return err
		}
		if len(result.Violations) > 0 {
			return exitError{code: 1}
		}
		return nil
	}

	if len(result.Violations) > 0 {
		if err := writeViolations(cmd, result.Violations); err != nil {
			return err
		}
		return exitError{code: 1}
	}

	_, err = io.WriteString(cmd.OutOrStdout(), result.Message)
	return err
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
