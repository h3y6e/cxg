package lint

import (
	"errors"
	"fmt"
	"io"

	"github.com/h3y6e/cxg/internal/core/message"
)

var ErrNoInput = errors.New("lint: no input")

type Input struct {
	Messages []string
	FilePath string
	Stdin    io.Reader
	Trailers []string
	Fix      bool
}

type Result struct {
	Message    string
	Violations []message.Violation
}

type Service struct {
	readFile ReadFile
}

func New(readFile ReadFile) *Service {
	return &Service{readFile: readFile}
}

func (s *Service) Lint(input Input) (Result, error) {
	value, err := s.resolve(input)
	if err != nil {
		return Result{}, err
	}

	value = message.AppendTrailers(value, input.Trailers)
	if input.Fix {
		value = message.Fix(value)
	}

	return Result{
		Message:    value,
		Violations: message.Validate(value),
	}, nil
}

func (s *Service) resolve(input Input) (string, error) {
	switch {
	case len(input.Messages) > 0:
		return message.Join(input.Messages), nil
	case input.Stdin != nil:
		content, err := io.ReadAll(input.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}

		return string(content), nil
	case input.FilePath != "":
		content, err := s.readFile(input.FilePath)
		if err != nil {
			return "", fmt.Errorf("reading file %q: %w", input.FilePath, err)
		}

		return string(content), nil
	default:
		return "", ErrNoInput
	}
}
