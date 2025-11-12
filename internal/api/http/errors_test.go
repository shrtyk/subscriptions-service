package http

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
	"github.com/stretchr/testify/assert"
)

func TestHttpError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		httpError *HttpError
		want      string
	}{
		{
			name: "With internal error",
			httpError: &HttpError{
				DTOErr: dto.Error{Code: 400, Message: "Bad Request"},
				err:    errors.New("validation failed"),
			},
			want: "http error: code=400, msg=Bad Request, internal_err=validation failed",
		},
		{
			name: "Without internal error",
			httpError: &HttpError{
				DTOErr: dto.Error{Code: 500, Message: "Internal Server Error"},
				err:    nil,
			},
			want: "http error: code=500, msg=Internal Server Error",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.httpError.Error())
		})
	}
}

func TestHttpError_Unwrap(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("internal error")
	httpError := &HttpError{
		DTOErr: dto.Error{Code: 500, Message: "Internal Server Error"},
		err:    internalErr,
	}

	assert.Equal(t, internalErr, httpError.Unwrap())
}

func TestNewHTTPError(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("something went wrong")
	httpError := NewHTTPError(http.StatusTeapot, "I'm a teapot", internalErr)

	assert.Equal(t, int32(http.StatusTeapot), httpError.DTOErr.Code)
	assert.Equal(t, "I'm a teapot", httpError.DTOErr.Message)
	assert.Equal(t, internalErr, httpError.err)
}

func TestBadRequestError(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("bad json")
	httpError := BadRequestError(internalErr)

	assert.Equal(t, int32(http.StatusBadRequest), httpError.DTOErr.Code)
	assert.Equal(t, "Badly formed request body", httpError.DTOErr.Message)
	assert.Equal(t, internalErr, httpError.err)
}

func TestInternalError(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("db connection failed")
	httpError := InternalError(internalErr)

	assert.Equal(t, int32(http.StatusInternalServerError), httpError.DTOErr.Code)
	assert.Equal(t, "Internal error", httpError.DTOErr.Message)
	assert.Equal(t, internalErr, httpError.err)
}

func Test_processAppError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want *HttpError
	}{
		{
			name: "DTOValidationError",
			err:  &DTOValidationError{ClientMessage: "Invalid input", InternalError: errors.New("field X missing")},
			want: NewHTTPError(http.StatusUnprocessableEntity, "Invalid input", &DTOValidationError{ClientMessage: "Invalid input", InternalError: errors.New("field X missing")}),
		},
		{
			name: "Service Error - KindNotFound",
			err:  subservice.NewErr("test", subservice.KindNotFound),
			want: NewHTTPError(http.StatusNotFound, "The requested resource was not found", subservice.NewErr("test", subservice.KindNotFound)),
		},
		{
			name: "Service Error - KindBusinessLogic",
			err:  subservice.NewErr("test", subservice.KindBusinessLogic),
			want: NewHTTPError(http.StatusUnprocessableEntity, "The operation cannot be completed due to a business rule violation", subservice.NewErr("test", subservice.KindBusinessLogic)),
		},
		{
			name: "Service Error - KindUnknown",
			err:  subservice.NewErr("test", subservice.KindUnknown),
			want: InternalError(subservice.NewErr("test", subservice.KindUnknown)),
		},
		{
			name: "Generic error",
			err:  errors.New("some generic error"),
			want: InternalError(errors.New("some generic error")),
		},
		{
			name: "Wrapped generic error",
			err:  fmt.Errorf("wrapped error: %w", errors.New("original error")),
			want: InternalError(fmt.Errorf("wrapped error: %w", errors.New("original error"))),
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := processAppError(tc.err)
			assert.Equal(t, tc.want.DTOErr.Code, got.DTOErr.Code)
			assert.Equal(t, tc.want.DTOErr.Message, got.DTOErr.Message)
			if tc.want.Unwrap() != nil {
				assert.EqualError(t, got.Unwrap(), tc.want.Unwrap().Error())
			} else {
				assert.Nil(t, got.Unwrap())
			}
		})
	}
}
