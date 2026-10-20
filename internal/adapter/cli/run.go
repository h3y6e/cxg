package cli

import (
	"encoding/json"
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

type failureResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}

func Run(options Options, linter linter) int {
	opts := &rootOptions{}
	root := newRootCmd(opts, options.Version, linter)

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

	command, err := root.ExecuteC()
	if err == nil {
		return 0
	}

	exitErr, ok := errors.AsType[exitError](err)
	if !ok {
		exitErr = exitError{status: 1, cause: err}
	}

	if exitErr.cause == nil {
		return exitErr.status
	}

	failure := failureResponse{Code: "failure", Message: exitErr.cause.Error()}
	if exitErr.status == 2 {
		failure.Code = "usage"
		failure.Fix = command.CommandPath() + " --help"
	}

	if opts.json {
		_ = json.NewEncoder(root.OutOrStdout()).Encode(lintResponse{
			Errors: []lintViolationResponse{},
			Error:  &failure,
		})

		return exitErr.status
	}

	_, _ = fmt.Fprintln(root.ErrOrStderr(), failure.Message)
	if failure.Fix != "" {
		_, _ = fmt.Fprintf(root.ErrOrStderr(), "Run '%s' for usage.\n", failure.Fix)
	}

	return exitErr.status
}
