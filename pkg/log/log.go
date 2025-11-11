package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
)

const (
	devEnv  = "dev"
	staging = "staging"
	prodEnv = "prod"
)

var fallbackLogger = slog.New(
	slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}),
)

func MustCreateNewLogger(env string) (log *slog.Logger) {
	switch env {
	case devEnv, staging:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}),
		)
	case prodEnv:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	default:
		msg := fmt.Sprintf("failed to create new logger. Unknown environment: '%s'", env)
		panic(msg)
	}
	return
}

type ctxKey string

const logCtxKey ctxKey = "logger"

func ToCtx(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, logCtxKey, log)
}

func FromCtx(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(logCtxKey).(*slog.Logger)
	if !ok {
		fallbackLogger.Error("failed to extract application logger out of context. Using fallback option.")
		return fallbackLogger
	}
	return l
}

func WithErr(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func NewTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := slog.New(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
	return l, &buf
}
