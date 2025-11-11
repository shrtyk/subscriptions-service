package http

import "fmt"

type DTOValidationError struct {
	ClientMessage string
	InternalError error
}

func (e *DTOValidationError) Error() string {
	if e.InternalError != nil {
		return fmt.Sprintf("%s: %v", e.ClientMessage, e.InternalError)
	}
	return e.ClientMessage
}

func (e *DTOValidationError) Unwrap() error {
	return e.InternalError
}
