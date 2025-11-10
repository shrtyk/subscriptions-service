package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
}
