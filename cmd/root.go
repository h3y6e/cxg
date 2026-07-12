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

func newRootCmd(version string, runner Runner) *cobra.Command {
	opts := &rootOptions{}

	root := &cobra.Command{
		Use:           "cxg",
		Short:         "Lint contextual messages for AI agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.PersistentFlags().BoolVar(&opts.json, "json", false, "Output machine-readable JSON")
	root.AddCommand(newLintCmd(opts, runner))

	return root
}
