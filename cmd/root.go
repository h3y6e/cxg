package cmd

import (
	"github.com/h3y6e/cxg/internal/lint"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	json bool
}

type Runner interface {
	Run(lint.Input) (lint.Result, error)
}

func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return usageError(err)
		}
		return nil
	}
}

func newRootCmd(version string, runner Runner) *cobra.Command {
	opts := &rootOptions{}

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
	root.AddCommand(newLintCmd(opts, runner))

	return root
}
