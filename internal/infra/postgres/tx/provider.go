package tx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/tx"
	"github.com/shrtyk/subscriptions-service/internal/infra/postgres"
)

type unitOfWork struct {
	tx       *sql.Tx
	repoCfg  *config.RepoConfig
	subsRepo repos.SubscriptionRepository
}

func (uow *unitOfWork) Subscriptions() repos.SubscriptionRepository {
	if uow.subsRepo == nil {
		uow.subsRepo = postgres.NewSubsRepo(uow.tx, uow.repoCfg)
	}
	return uow.subsRepo
}

type provider struct {
	db      *sql.DB
	repoCfg *config.RepoConfig
}

func NewProvider(db *sql.DB, repoCfg *config.RepoConfig) tx.Provider {
	return &provider{
		db:      db,
		repoCfg: repoCfg,
	}
}

func (p *provider) WithTransaction(ctx context.Context, txFunc func(uow tx.UnitOfWork) error) error {
	sqlTx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// If panic occurred call rollback and panic again
	defer func() {
		if r := recover(); r != nil {
			_ = sqlTx.Rollback()
			panic(r)
		}
	}()

	uow := &unitOfWork{
		tx:      sqlTx,
		repoCfg: p.repoCfg,
	}

	if err = txFunc(uow); err != nil {
		if rollbackErr := sqlTx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction rollback failed after function error: %v (original error: %w)", rollbackErr, err)
		}
		return err
	}

	if err = sqlTx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
