package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	cobrausage "github.com/jdx/usage/integrations/cobra"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	json      bool
	usageSpec bool
	terminal  bool
	envErr    *exitError
}

func (o rootOptions) checkEnv() error {
	if o.envErr != nil {
		return *o.envErr
	}

	return nil
}

func newRootCmd(opts *rootOptions, version string, linter linter) *cobra.Command {
	root := &cobra.Command{
		Use:                        "cxg",
		Short:                      "Lint contextual messages for AI agents",
		SilenceUsage:               true,
		SilenceErrors:              true,
		SuggestionsMinimumDistance: 2,
		Version:                    version,
		Args:                       rootArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.usageSpec {
				_, err := io.WriteString(cmd.OutOrStdout(), cobrausage.Generate(cmd))
				return err
			}

			return cmd.Help()
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return exitError{status: 2, cause: err, unparsed: true}
	})

	root.PersistentFlags().BoolVar(&opts.json, "json", false, "Output machine-readable JSON")
	root.Flags().BoolVar(&opts.usageSpec, "usage-spec", false, "Print the usage spec")
	_ = root.Flags().MarkHidden("usage-spec")
	root.AddCommand(newLintCmd(opts, linter))
	documentEnv(root)

	return root
}

func rootArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	message := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		message += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}

	return usageError(errors.New(message))
}

func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return usageError(err)
		}

		return nil
	}
}
