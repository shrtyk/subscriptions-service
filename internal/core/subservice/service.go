package subservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/tx"
	"github.com/shrtyk/subscriptions-service/pkg/errkit"
)

const (
	opCreate    = "subservice.Create"
	opGetByID   = "subservice.GetByID"
	opUpdate    = "subservice.Update"
	opDelete    = "subservice.Delete"
	opList      = "subservice.List"
	opTotalCost = "subservice.TotalCost"
)

type service struct {
	repo       repos.SubscriptionRepository
	txProvider tx.Provider
}

func New(repo repos.SubscriptionRepository, txProvider tx.Provider) subservice.SubscriptionsService {
	return &service{
		repo:       repo,
		txProvider: txProvider,
	}
}

func (s *service) Create(ctx context.Context, sub domain.Subscription) (*domain.Subscription, error) {
	err := s.repo.Create(ctx, &sub)
	if err != nil {
		var repoErr *errkit.BaseErr[repos.RepoKind]
		if errors.As(err, &repoErr) && repoErr.Kind == repos.KindDuplicate {
			return nil, subservice.WrapErr(opCreate, subservice.KindBusinessLogic, err)
		}
		return nil, subservice.WrapErr(opCreate, subservice.KindUnknown, err)
	}

	return &sub, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		var repoErr *errkit.BaseErr[repos.RepoKind]
		if errors.As(err, &repoErr) && repoErr.Kind == repos.KindNotFound {
			return nil, subservice.WrapErr(opGetByID, subservice.KindNotFound, err)
		}
		return nil, subservice.WrapErr(opGetByID, subservice.KindUnknown, err)
	}
	return sub, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, update domain.SubscriptionUpdate) (*domain.Subscription, error) {
	var updatedSub *domain.Subscription
	err := s.txProvider.WithTransaction(ctx, func(uow tx.UnitOfWork) error {
		repo := uow.Subscriptions()

		existing, err := repo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if update.ServiceName != nil {
			existing.ServiceName = *update.ServiceName
		}
		if update.MonthlyCost != nil {
			existing.MonthlyCost = *update.MonthlyCost
		}
		if update.ClearEndDate {
			existing.EndDate = nil
		} else if update.EndDate != nil {
			existing.EndDate = update.EndDate
		}
		existing.UpdatedAt = time.Now().UTC()

		if existing.EndDate != nil && existing.StartDate.After(*existing.EndDate) {
			return subservice.WrapErr(opUpdate, subservice.KindBusinessLogic, errors.New("end_date cannot be before start_date"))
		}

		if err := repo.Update(ctx, existing); err != nil {
			return err
		}
		updatedSub = existing
		return nil
	})

	if err != nil {
		var repoErr *errkit.BaseErr[repos.RepoKind]
		if errors.As(err, &repoErr) {
			switch repoErr.Kind {
			case repos.KindNotFound:
				return nil, subservice.WrapErr(opUpdate, subservice.KindNotFound, err)
			case repos.KindDuplicate:
				return nil, subservice.WrapErr(opUpdate, subservice.KindBusinessLogic, err)
			}
		}
		var serviceErr *errkit.BaseErr[subservice.ServiceKind]
		if errors.As(err, &serviceErr) && serviceErr.Kind == subservice.KindBusinessLogic {
			return nil, err
		}
		return nil, subservice.WrapErr(opUpdate, subservice.KindUnknown, err)
	}

	return updatedSub, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		var repoErr *errkit.BaseErr[repos.RepoKind]
		if errors.As(err, &repoErr) && repoErr.Kind == repos.KindNotFound {
			return subservice.WrapErr(opDelete, subservice.KindNotFound, err)
		}
		return subservice.WrapErr(opDelete, subservice.KindUnknown, err)
	}
	return nil
}

func (s *service) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, subservice.WrapErr(opList, subservice.KindUnknown, err)
	}
	return subs, nil
}

func (s *service) TotalCost(ctx context.Context, filter domain.SubscriptionFilter, start, end time.Time) (int, error) {
	subs, err := s.repo.ListAll(ctx, filter)
	if err != nil {
		return 0, subservice.WrapErr(opTotalCost, subservice.KindUnknown, err)
	}

	totalCost := 0
	for _, sub := range subs {
		totalCost += calculateCostForSubscription(sub, start, end)
	}

	return totalCost, nil
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func calculateCostForSubscription(sub domain.Subscription, periodStart, periodEnd time.Time) int {
	subStart := startOfMonth(sub.StartDate)
	pStart := startOfMonth(periodStart)
	pEnd := startOfMonth(periodEnd)

	if subStart.After(pEnd) {
		return 0
	}

	if sub.EndDate != nil {
		subEnd := startOfMonth(*sub.EndDate)
		if subEnd.Before(pStart) {
			return 0
		}
	}

	billingStart := pStart
	if subStart.After(pStart) {
		billingStart = subStart
	}

	billingEnd := pEnd
	if sub.EndDate != nil {
		subEnd := startOfMonth(*sub.EndDate)
		if subEnd.Before(pEnd) {
			billingEnd = subEnd
		}
	}

	if billingStart.After(billingEnd) {
		return 0
	}

	months := (billingEnd.Year()-billingStart.Year())*12 +
		int(billingEnd.Month()) - int(billingStart.Month()) + 1

	return months * sub.MonthlyCost
}
