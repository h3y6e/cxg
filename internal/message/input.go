package message

import (
	"errors"
	"io"
	"os"
	"strings"
)

var ErrNoInput = errors.New("no commit message input provided")

type Input struct {
	Messages []string
	FilePath string
	Stdin    io.Reader
	HasStdin bool
	Trailers []string
}

func Resolve(input Input) (string, error) {
	switch {
	case len(input.Messages) > 0:
		return appendTrailers(joinMessageFlags(input.Messages), input.Trailers), nil
	case input.HasStdin:
		content, err := io.ReadAll(input.Stdin)
		if err != nil {
			return "", err
		}

		return appendTrailers(trimFinalNewlines(string(content)), input.Trailers), nil
	case input.FilePath != "":
		content, err := os.ReadFile(input.FilePath)
		if err != nil {
			return "", err
		}

		return appendTrailers(trimFinalNewlines(string(content)), input.Trailers), nil
	default:
		return "", ErrNoInput
	}
}

func joinMessageFlags(messages []string) string {
	if len(messages) == 1 {
		return messages[0]
	}

	return messages[0] + "\n\n" + strings.Join(messages[1:], "\n")
}

func appendTrailers(message string, trailers []string) string {
	if len(trailers) == 0 {
		return message
	}

	if message == "" {
		return strings.Join(trailers, "\n")
	}

	return message + "\n\n" + strings.Join(trailers, "\n")
}

func trimFinalNewlines(value string) string {
	return strings.TrimRight(value, "\r\n")
}
