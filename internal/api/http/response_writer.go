package http

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"

	"github.com/shrtyk/subscriptions-service/pkg/log"
)

type HttpResponse[T any] struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Body    T      `json:"data,omitempty"`
}

type customResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func WriteHTTPError(w http.ResponseWriter, r *http.Request, e *HttpError) {
	l := log.FromCtx(r.Context())
	if e.DTOErr.Code >= 500 {
		l.Error("Server error", log.WithErr(e))
	} else {
		l.Info("Client error", log.WithErr(e))
	}

	err := WriteJSON(
		w,
		e.DTOErr,
		int(e.DTOErr.Code),
		nil,
	)
	if err != nil {
		l.Error("Failed to response with error", log.WithErr(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func WriteJSON[T any](w http.ResponseWriter, data T, status int, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Uncomment to add extra readability during manual testing:
	//
	b = append(b, '\n')
	buf := bytes.Buffer{}
	if err := json.Indent(&buf, b, "", "\t"); err != nil {
		return err
	}
	b = buf.Bytes()
	//

	maps.Copy(w.Header(), headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(b); err != nil {
		return err
	}

	return nil
}
