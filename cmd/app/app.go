package main

import (
	"log/slog"

	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
)

type application struct {
	Cfg         *config.Config
	Logger      *slog.Logger
	SubsRepo    repos.SubscriptionRepository
	SubsService subservice.SubscriptionsService
}

type option func(*application)

func NewApplication(opts ...option) *application {
	app := &application{}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func WithConfig(cfg *config.Config) option {
	return func(app *application) {
		app.Cfg = cfg
	}
}

func WithLogger(l *slog.Logger) option {
	return func(app *application) {
		app.Logger = l
	}
}

func WithRepo(r repos.SubscriptionRepository) option {
	return func(app *application) {
		app.SubsRepo = r
	}
}

func WithSubsService(s subservice.SubscriptionsService) option {
	return func(app *application) {
		app.SubsService = s
	}
}
