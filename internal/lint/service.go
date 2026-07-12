package lint

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrNoInput = errors.New("lint: no input")

type ReadFile func(string) ([]byte, error)

type Input struct {
	Messages []string
	FilePath string
	Stdin    io.Reader
	Trailers []string
	Fix      bool
}

type Result struct {
	Message    string
	Violations []Violation
}

type Violation struct {
	Line    int
	Code    string
	Message string
}

type Service struct {
	readFile ReadFile
}

func New(readFile ReadFile) *Service {
	return &Service{readFile: readFile}
}

func (s *Service) Run(input Input) (Result, error) {
	value, err := s.resolve(input)
	if err != nil {
		return Result{}, err
	}

	if input.Fix {
		value = fix(value)
	}

	return Result{
		Message:    value,
		Violations: validate(value),
	}, nil
}

func (s *Service) resolve(input Input) (string, error) {
	var value string
	switch {
	case len(input.Messages) > 0:
		value = joinMessages(input.Messages)
	case input.Stdin != nil:
		content, err := io.ReadAll(input.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		value = trimFinalNewlines(string(content))
	case input.FilePath != "":
		content, err := s.readFile(input.FilePath)
		if err != nil {
			return "", fmt.Errorf("reading file %q: %w", input.FilePath, err)
		}
		value = trimFinalNewlines(string(content))
	default:
		return "", ErrNoInput
	}

	return appendTrailers(value, input.Trailers), nil
}

func joinMessages(messages []string) string {
	if len(messages) == 1 {
		return messages[0]
	}

	return messages[0] + "\n\n" + strings.Join(messages[1:], "\n")
}

func appendTrailers(value string, trailers []string) string {
	if len(trailers) == 0 {
		return value
	}
	if value == "" {
		return strings.Join(trailers, "\n")
	}

	return value + "\n\n" + strings.Join(trailers, "\n")
}

func trimFinalNewlines(value string) string {
	return strings.TrimRight(value, "\r\n")
}
