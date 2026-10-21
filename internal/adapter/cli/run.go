package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/h3y6e/cxg/internal/core/lint"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Options struct {
	Version       string
	Args          []string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	Env           []string
	StdinTerminal bool
}

type exitError struct {
	status   int
	cause    error
	fix      string
	unparsed bool
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
	opts := &rootOptions{terminal: options.StdinTerminal}
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

	opts.envErr = bindEnv(root, args, options.Env)

	command, err := root.ExecuteC()
	if err == nil {
		return 0
	}

	exitErr, ok := errors.AsType[exitError](err)
	if !ok {
		exitErr = exitError{status: 1, cause: err}
	}

	if exitErr.cause != nil {
		report(command, failureFor(command, exitErr), opts.json && !exitErr.unparsed)
	}

	return exitErr.status
}

func failureFor(command *cobra.Command, exitErr exitError) failureResponse {
	if exitErr.status == 2 {
		fix := exitErr.fix
		if fix == "" {
			fix = command.CommandPath() + " --help"
		}

		return failureResponse{Code: "usage", Message: exitErr.cause.Error(), Fix: fix}
	}

	return failureResponse{Code: "failure", Message: exitErr.cause.Error()}
}

func report(command *cobra.Command, failure failureResponse, asJSON bool) {
	if asJSON {
		var record any = newLintResponse(lint.Result{}, &failure)
		if command.Name() == "recall" {
			record = newRecallResponse(nil, &failure)
		}

		if err := json.NewEncoder(command.OutOrStdout()).Encode(record); err == nil {
			return
		}
	}

	_, _ = fmt.Fprintln(command.ErrOrStderr(), failure.Message)
	if failure.Fix != "" {
		_, _ = fmt.Fprintf(command.ErrOrStderr(), "Fix: %s\n", failure.Fix)
	}
}

func bindEnv(root *cobra.Command, args, env []string) *exitError {
	values := map[string]string{}

	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		values[name] = value
	}

	command, _, err := root.Find(args)
	if err != nil {
		command = root
	}

	var failure *exitError

	bind := func(flag *pflag.Flag) {
		name, ok := envName(flag)
		if !ok {
			return
		}

		value, ok := values[name]
		if !ok {
			return
		}

		if err := flag.Value.Set(value); err != nil && failure == nil {
			failure = &exitError{status: 2, cause: fmt.Errorf("invalid %s: %w", name, err), fix: "unset " + name}
		}
	}

	command.LocalFlags().VisitAll(bind)
	command.InheritedFlags().VisitAll(bind)

	return failure
}

func documentEnv(root *cobra.Command) {
	document := func(flag *pflag.Flag) {
		if name, ok := envName(flag); ok {
			flag.Usage += " [$" + name + "]"
		}
	}

	root.PersistentFlags().VisitAll(document)

	for _, command := range root.Commands() {
		command.Flags().VisitAll(document)
	}
}

func envName(flag *pflag.Flag) (string, bool) {
	if flag.Hidden || flag.Value.Type() == "stringArray" {
		return "", false
	}

	return "CXG_" + strings.ToUpper(strings.ReplaceAll(flag.Name, "-", "_")), true
}
