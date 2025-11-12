package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/core/subservice"
	"github.com/shrtyk/subscriptions-service/internal/infra/postgres"
	"github.com/shrtyk/subscriptions-service/internal/infra/postgres/tx"
	"github.com/shrtyk/subscriptions-service/pkg/log"
)

func main() {
	cfg := config.MustInitConfig()
	l := log.MustCreateNewLogger(cfg.AppCfg.Env)
	db := postgres.MustCreateConnectionPool(&cfg.PostgresCfg)

	subsRepo := postgres.NewSubsRepo(db, &cfg.RepoCfg)
	txProvider := tx.NewProvider(db, &cfg.RepoCfg)
	subsService := subservice.New(subsRepo, txProvider)

	app := NewApplication(
		WithConfig(cfg),
		WithLogger(l),
		WithRepo(subsRepo),
		WithSubsService(subsService),
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	app.Serve(ctx)
}
