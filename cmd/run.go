package cmd

import (
	"errors"
	"fmt"
	"io"
)

type Options struct {
	Version string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

type exitError struct {
	code int
}

func (err exitError) Error() string {
	return fmt.Sprintf("exit %d", err.code)
}

func Run(options Options, runner Runner) int {
	command := newRootCmd(options.Version, runner)
	command.SetArgs(options.Args)
	if options.Stdin != nil {
		command.SetIn(options.Stdin)
	}
	if options.Stdout != nil {
		command.SetOut(options.Stdout)
	}
	if options.Stderr != nil {
		command.SetErr(options.Stderr)
	}

	err := command.Execute()
	if err == nil {
		return 0
	}

	var exitError exitError
	if errors.As(err, &exitError) {
		return exitError.code
	}

	_, _ = fmt.Fprintln(command.ErrOrStderr(), err)
	return 1
}
