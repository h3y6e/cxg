package cli

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
	status int
	cause  error
}

func (err exitError) Error() string {
	if err.cause != nil {
		return err.cause.Error()
	}

	return fmt.Sprintf("exit %d", err.status)
}

func (err exitError) Unwrap() error {
	return err.cause
}

func usageError(err error) error {
	return exitError{status: 2, cause: err}
}

func Run(options Options, linter linter) int {
	root := newRootCmd(&rootOptions{}, options.Version, linter)

	args := options.Args
	if args == nil {
		args = []string{}
	}

	root.SetArgs(args)

	if options.Stdin != nil {
		root.SetIn(options.Stdin)
	}

	if options.Stdout != nil {
		root.SetOut(options.Stdout)
	}

	if options.Stderr != nil {
		root.SetErr(options.Stderr)
	}

	err := root.Execute()
	if err == nil {
		return 0
	}

	exitErr, ok := errors.AsType[exitError](err)
	if !ok {
		exitErr = exitError{status: 1, cause: err}
	}

	if exitErr.cause != nil {
		_, _ = fmt.Fprintln(root.ErrOrStderr(), exitErr.cause)
	}

	return exitErr.status
}
