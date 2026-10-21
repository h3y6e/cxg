package main

import (
	"os"

	"github.com/h3y6e/cxg/internal/adapter/cli"
	"github.com/h3y6e/cxg/internal/core/lint"
	"golang.org/x/term"
)

var version = "dev"

func main() {
	code := cli.Run(cli.Options{
		Version:       cli.Version(version),
		Args:          os.Args[1:],
		Stdin:         os.Stdin,
		Stdout:        os.Stdout,
		Stderr:        os.Stderr,
		Env:           os.Environ(),
		StdinTerminal: term.IsTerminal(int(os.Stdin.Fd())),
	}, lint.New(os.ReadFile))
	if code != 0 {
		os.Exit(code)
	}
}
