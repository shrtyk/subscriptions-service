package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tomasen/realip"
)

const maxRequestBodyBytes = 1_048_576

func ReadJSON[T any](w http.ResponseWriter, r *http.Request, dst T) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxRequestBodyBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			// Shouldn't occur at all. Only possible if wrong value passed as dst
			panic(err)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != nil {
		if !errors.Is(err, io.EOF) {
			return errors.New("body must contain a single JSON value")
		}
	}

	return nil
}

func SubIdParam(r *http.Request) (*uuid.UUID, error) {
	strId := chi.URLParam(r, "id")
	if err := uuid.Validate(strId); err != nil {
		return nil, err
	}

	subId, err := uuid.Parse(strId)
	if err != nil {
		return nil, err
	}

	return &subId, nil
}

func UserAgentAndIP(r *http.Request) (string, string) {
	return r.UserAgent(), realip.FromRequest(r)
}
