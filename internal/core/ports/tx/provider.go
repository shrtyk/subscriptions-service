package tx

import (
	"context"

	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
)

//go:generate mockery
type UnitOfWork interface {
	Subscriptions() repos.SubscriptionRepository
}

//go:generate mockery
type Provider interface {
	WithTransaction(ctx context.Context, fn func(uow UnitOfWork) error) error
}
