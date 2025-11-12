package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
	"github.com/shrtyk/subscriptions-service/pkg/errkit"
)

type HttpError struct {
	DTOErr dto.Error
	err    error
}

func (e *HttpError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("http error: code=%d, msg=%s, internal_err=%v", e.DTOErr.Code, e.DTOErr.Message, e.err)
	}
	return fmt.Sprintf("http error: code=%d, msg=%s", e.DTOErr.Code, e.DTOErr.Message)
}

func (e *HttpError) Unwrap() error {
	return e.err
}

func NewHTTPError(code int, message string, internalErr error) *HttpError {
	return &HttpError{
		DTOErr: dto.Error{
			Code:    int32(code),
			Message: message,
		},
		err: internalErr,
	}
}

func BadRequestError(err error) *HttpError {
	return &HttpError{
		DTOErr: dto.Error{
			Code:    http.StatusBadRequest,
			Message: "Badly formed request body",
		},
		err: err,
	}
}

func InternalError(err error) *HttpError {
	return &HttpError{
		DTOErr: dto.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal error",
		},
		err: err,
	}
}

func processAppError(err error) *HttpError {
	var dtoErr *DTOValidationError
	if errors.As(err, &dtoErr) {
		return NewHTTPError(http.StatusUnprocessableEntity, dtoErr.ClientMessage, dtoErr)
	}

	var serviceErr *errkit.BaseErr[subservice.ServiceKind]
	if errors.As(err, &serviceErr) {
		switch serviceErr.Kind {
		case subservice.KindNotFound:
			return NewHTTPError(http.StatusNotFound, "The requested resource was not found", serviceErr)
		case subservice.KindBusinessLogic:
			return NewHTTPError(http.StatusUnprocessableEntity, "The operation cannot be completed due to a business rule violation", serviceErr)
		default:
			return InternalError(serviceErr)
		}
	}

	return InternalError(err)
}
