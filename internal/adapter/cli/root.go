package cli

import "github.com/spf13/cobra"

type rootOptions struct {
	json bool
}

func newRootCmd(opts *rootOptions, version string, linter linter) *cobra.Command {
	root := &cobra.Command{
		Use:           "cxg",
		Short:         "Lint contextual messages for AI agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		Args:          usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError(err)
	})

	root.PersistentFlags().BoolVar(&opts.json, "json", false, "Output machine-readable JSON")
	root.AddCommand(newLintCmd(opts, linter))

	return root
}

func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return usageError(err)
		}

		return nil
	}
}
