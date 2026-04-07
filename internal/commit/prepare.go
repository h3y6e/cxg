package commit

import (
	"io"

	cxglint "github.com/h3y6e/cxg/internal/lint"
	"github.com/h3y6e/cxg/internal/message"
)

type PrepareRequest struct {
	Messages []string
	FilePath string
	Stdin    io.Reader
	HasStdin bool
	Trailers []string
	Fix      bool
}

type PreparedMessage struct {
	Message string
	Errors  []message.ValidationError
}

func Prepare(request PrepareRequest) (PreparedMessage, error) {
	value, err := message.Resolve(message.Input{
		Messages: request.Messages,
		FilePath: request.FilePath,
		Stdin:    request.Stdin,
		HasStdin: request.HasStdin,
		Trailers: request.Trailers,
	})
	if err != nil {
		return PreparedMessage{}, err
	}

	if request.Fix {
		value = message.Fix(value)
	}

	return PreparedMessage{
		Message: value,
		Errors:  cxglint.Validate(value),
	}, nil
}
