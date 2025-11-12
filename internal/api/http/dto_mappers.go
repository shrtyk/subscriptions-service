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

func fromUpdateSubscriptionRequest(req *UpdateSubscriptionRequest) (*domain.SubscriptionUpdate, error) {
	endDate, clearEndDate, err := validateUpdateSubscriptionRequest(req)
	if err != nil {
		return nil, err
	}

	domainUpdate := &domain.SubscriptionUpdate{
		ServiceName:  req.ServiceName,
		MonthlyCost:  req.MonthlyCost,
		EndDate:      endDate,
		ClearEndDate: clearEndDate,
	}

	return domainUpdate, nil
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
