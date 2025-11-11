package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/pkg/log"
)

type middlewares struct {
	log *slog.Logger
}

func NewMiddlewaresProvider(log *slog.Logger) *middlewares {
	return &middlewares{
		log: log,
	}
}

func (m middlewares) PanicRecoveryMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.log.Error(
					"Error occurred",
					log.WithErr(fmt.Errorf("%v", err)),
				)
				w.Header().Set("Connection", "close")
				WriteHTTPError(w, r, NewHTTPError(
					http.StatusInternalServerError,
					"The server encountered a problem and could not process your request",
					fmt.Errorf("%s", err)))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (m middlewares) LoggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua, ip := UserAgentAndIP(r)
		reqID := uuid.NewString()

		l := m.log.With(
			slog.String("ip", ip),
			slog.String("user_agent", ua),
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
		)

		l.Debug("New HTTP request")
		newCtx := log.ToCtx(r.Context(), l)
		newReq := r.WithContext(newCtx)
		custWriter := &customResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		reqStart := time.Now()
		next.ServeHTTP(custWriter, newReq)
		reqEnd := time.Since(reqStart)

		ttp := fmt.Sprintf("%.5fs", reqEnd.Seconds())
		l.Debug(
			"HTTP request processed",
			slog.Int("status_code", custWriter.statusCode),
			slog.String("request_duration", ttp),
		)
	})
}
