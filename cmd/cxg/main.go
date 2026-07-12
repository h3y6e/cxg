package main

import (
	"os"

	"github.com/h3y6e/cxg/cmd"
	"github.com/h3y6e/cxg/internal/lint"
)

var version = "dev"

func main() {
	code := cmd.Run(cmd.Options{
		Version: version,
		Args:    os.Args[1:],
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}, lint.New(os.ReadFile))
	if code != 0 {
		os.Exit(code)
	}
}
