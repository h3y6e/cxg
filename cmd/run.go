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
	code  int
	cause error
}

func (err exitError) Error() string {
	if err.cause != nil {
		return err.cause.Error()
	}
	return fmt.Sprintf("exit %d", err.code)
}

func (err exitError) Unwrap() error {
	return err.cause
}

func usageError(err error) error {
	return exitError{code: 2, cause: err}
}

func Run(options Options, runner Runner) int {
	command := newRootCmd(options.Version, runner)
	args := options.Args
	if args == nil {
		args = []string{}
	}
	command.SetArgs(args)
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

	if exitErr, ok := errors.AsType[exitError](err); ok {
		if exitErr.cause != nil {
			_, _ = fmt.Fprintln(command.ErrOrStderr(), exitErr.cause)
		}
		return exitErr.code
	}

	_, _ = fmt.Fprintln(command.ErrOrStderr(), err)
	return 1
}
