package commit

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	internalgit "github.com/h3y6e/cxg/internal/git"
	"github.com/h3y6e/cxg/internal/message"
)

type Request struct {
	PrepareRequest
	Dir        string
	All        bool
	Amend      bool
	AllowEmpty bool
	Signoff    bool
	Author     string
	Date       string
}

type Result struct {
	Valid     bool                      `json:"valid"`
	Committed bool                      `json:"committed"`
	Message   string                    `json:"message,omitempty"`
	Errors    []message.ValidationError `json:"errors"`
	GitError  *GitError                 `json:"gitError,omitempty"`
	Stdout    string                    `json:"-"`
	Stderr    string                    `json:"-"`
}

type GitError struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

func Run(ctx context.Context, request Request) (Result, error) {
	prepared, err := Prepare(request.PrepareRequest)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Valid:   len(prepared.Errors) == 0,
		Message: prepared.Message,
		Errors:  prepared.Errors,
	}
	if len(prepared.Errors) > 0 {
		return result, nil
	}

	cmd, err := internalgit.Command(ctx, request.Dir, buildCommitArgs(request)...)
	if err != nil {
		return Result{}, err
	}
	cmd.Stdin = strings.NewReader(prepared.Message)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.GitError = &GitError{
				ExitCode: exitErr.ExitCode(),
				Stdout:   result.Stdout,
				Stderr:   result.Stderr,
			}
			return result, nil
		}
		return Result{}, err
	}

	result.Committed = true
	return result, nil
}

func buildCommitArgs(request Request) []string {
	args := []string{"commit"}
	if request.All {
		args = append(args, "--all")
	}
	if request.Amend {
		args = append(args, "--amend")
	}
	if request.AllowEmpty {
		args = append(args, "--allow-empty")
	}
	if request.Signoff {
		args = append(args, "--signoff")
	}
	if request.Author != "" {
		args = append(args, "--author", request.Author)
	}
	if request.Date != "" {
		args = append(args, "--date", request.Date)
	}

	args = append(args, "-F", "-")
	return args
}
