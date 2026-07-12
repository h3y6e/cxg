package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/h3y6e/cxg/internal/lint"
	"github.com/spf13/cobra"
)

type lintResponse struct {
	Valid   bool                    `json:"valid"`
	Message string                  `json:"message,omitempty"`
	Errors  []lintViolationResponse `json:"errors"`
}

type lintViolationResponse struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeLintJSON(cmd *cobra.Command, result lint.Result) error {
	errors := make([]lintViolationResponse, 0, len(result.Violations))
	for _, violation := range result.Violations {
		errors = append(errors, lintViolationResponse{
			Line:    violation.Line,
			Code:    violation.Code,
			Message: violation.Message,
		})
	}

	response := lintResponse{
		Valid:  len(result.Violations) == 0,
		Errors: errors,
	}
	if response.Valid {
		response.Message = result.Message
	}

	return json.NewEncoder(cmd.OutOrStdout()).Encode(response)
}

func writeViolations(cmd *cobra.Command, violations []lint.Violation) error {
	for _, violation := range violations {
		_, err := fmt.Fprintf(
			cmd.ErrOrStderr(),
			"line %d [%s] %s\n",
			violation.Line,
			violation.Code,
			violation.Message,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
