package cmd

import "github.com/spf13/cobra"

type rootOptions struct {
	json bool
}

func Execute(version string) error {
	return NewRootCmd(version).Execute()
}

func NewRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}

	root := &cobra.Command{
		Use:           "cxg",
		Short:         "Lint contextual messages for AI agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.PersistentFlags().BoolVar(&opts.json, "json", false, "Output machine-readable JSON")
	root.AddCommand(newLintCmd(opts))

	return root
}
