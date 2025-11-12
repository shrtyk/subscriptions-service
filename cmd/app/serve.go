package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	appHttp "github.com/shrtyk/subscriptions-service/internal/api/http"
	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/pkg/log"
)

func (app *application) Serve(ctx context.Context) {
	mws := appHttp.NewMiddlewaresProvider(app.Logger)
	h := appHttp.NewHandler(app.SubsService)

	server := http.Server{
		Addr: ":" + app.Cfg.HttpCfg.Port,
		Handler: dto.HandlerWithOptions(h, dto.ChiServerOptions{
			BaseURL:     "/api/v1",
			Middlewares: []dto.MiddlewareFunc{mws.PanicRecoveryMW, mws.LoggingMW},
		}),
	}

	eChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		tctx, tcancel := context.WithTimeout(context.Background(), app.Cfg.AppCfg.ShutdownTimeout)
		defer tcancel()

		eChan <- server.Shutdown(tctx)
	}()

	app.Logger.Info(
		"HTTP server successfully started",
		slog.String("address", ":"+app.Cfg.HttpCfg.Port),
	)

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.Logger.Error("server failed", log.WithErr(err))
		return
	}

	if cerr := <-eChan; cerr != nil {
		app.Logger.Error("failed graceful shutdown", log.WithErr(cerr))
		return
	}

	app.Logger.Info("graceful shutdown completed successfully")
}
