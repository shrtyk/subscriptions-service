package subservice

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
)

//go:generate mockery
type SubscriptionsService interface {
	Create(ctx context.Context, sub domain.Subscription) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, sub domain.Subscription) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
	TotalCost(ctx context.Context, filter domain.SubscriptionFilter, start, end time.Time) (int, error)
}
