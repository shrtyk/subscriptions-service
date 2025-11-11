package http

import (
	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
)

func toSubscriptionDTO(sub *domain.Subscription) *dto.Subscription {
	var endDate *string
	if sub.EndDate != nil {
		s := dateLayout.format(*sub.EndDate)
		endDate = &s
	}

	return &dto.Subscription{
		Id:          sub.ID,
		ServiceName: sub.ServiceName,
		MonthlyCost: sub.MonthlyCost,
		UserId:      sub.UserID,
		StartDate:   dateLayout.format(sub.StartDate),
		EndDate:     endDate,
	}
}

func fromNewSubscriptionDTO(d *dto.NewSubscription) (*domain.Subscription, error) {
	startDate, endDate, err := validateNewSubscription(d)
	if err != nil {
		return nil, err
	}

	return &domain.Subscription{
		ServiceName: d.ServiceName,
		MonthlyCost: d.MonthlyCost,
		UserID:      uuid.UUID(d.UserId),
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func fromUpdateSubscriptionDTO(d *dto.UpdateSubscription) (*domain.SubscriptionUpdate, error) {
	endDate, err := validateUpdateSubscription(d)
	if err != nil {
		return nil, err
	}

	return &domain.SubscriptionUpdate{
		ServiceName: d.ServiceName,
		MonthlyCost: d.MonthlyCost,
		EndDate:     endDate,
	}, nil
}

func toListFilter(params dto.ListSubscriptionsParams) domain.SubscriptionFilter {
	filter := domain.SubscriptionFilter{
		Page:        params.Page,
		PageSize:    params.PageSize,
		ServiceName: params.ServiceName,
	}
	if params.UserId != nil {
		uid := uuid.UUID(*params.UserId)
		filter.UserID = &uid
	}
	return filter
}
