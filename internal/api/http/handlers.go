package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
)

type handler struct {
	service subservice.SubscriptionsService
}

func NewHandler(service subservice.SubscriptionsService) *handler {
	return &handler{service: service}
}

func (h *handler) ListSubscriptions(w http.ResponseWriter, r *http.Request, params dto.ListSubscriptionsParams) {
	filter := toListFilter(params)

	subs, err := h.service.List(r.Context(), filter)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	dtoSubs := make([]dto.Subscription, 0, len(subs))
	for _, sub := range subs {
		dtoSubs = append(dtoSubs, *toSubscriptionDTO(&sub))
	}

	if err := WriteJSON(w, dtoSubs, http.StatusOK, nil); err != nil {
		WriteHTTPError(w, r, processAppError(err))
	}
}

func (h *handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var body dto.NewSubscription
	if err := ReadJSON(w, r, &body); err != nil {
		WriteHTTPError(w, r, BadRequestError(err))
		return
	}

	domainSub, err := fromNewSubscriptionDTO(&body)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	createdSub, err := h.service.Create(r.Context(), *domainSub)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	if err := WriteJSON(w, toSubscriptionDTO(createdSub), http.StatusCreated, nil); err != nil {
		WriteHTTPError(w, r, processAppError(err))
	}
}

func (h *handler) GetTotalCost(w http.ResponseWriter, r *http.Request, params dto.GetTotalCostParams) {
	filter := domain.SubscriptionFilter{}
	if params.UserId != nil {
		uid := uuid.UUID(*params.UserId)
		filter.UserID = &uid
	}
	if params.ServiceName != nil {
		filter.ServiceName = params.ServiceName
	}

	start, end, valErr := validateGetTotalCostParams(params)
	if valErr != nil {
		WriteHTTPError(w, r, processAppError(valErr))
		return
	}

	total, err := h.service.TotalCost(r.Context(), filter, start, end)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	response := struct {
		TotalCost int `json:"total_cost"`
	}{
		TotalCost: total,
	}

	if err := WriteJSON(w, response, http.StatusOK, nil); err != nil {
		WriteHTTPError(w, r, processAppError(err))
	}
}

func (h *handler) DeleteSubscription(w http.ResponseWriter, r *http.Request, id types.UUID) {
	if err := h.service.Delete(r.Context(), uuid.UUID(id)); err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) GetSubscriptionById(w http.ResponseWriter, r *http.Request, id types.UUID) {
	sub, err := h.service.GetByID(r.Context(), uuid.UUID(id))
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	if err := WriteJSON(w, toSubscriptionDTO(sub), http.StatusOK, nil); err != nil {
		WriteHTTPError(w, r, processAppError(err))
	}
}

func (h *handler) UpdateSubscription(w http.ResponseWriter, r *http.Request, id types.UUID) {
	var body dto.UpdateSubscription
	if err := ReadJSON(w, r, &body); err != nil {
		WriteHTTPError(w, r, BadRequestError(err))
		return
	}

	domainUpdate, err := fromUpdateSubscriptionDTO(&body)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	updatedSub, err := h.service.Update(r.Context(), uuid.UUID(id), *domainUpdate)
	if err != nil {
		WriteHTTPError(w, r, processAppError(err))
		return
	}

	if err := WriteJSON(w, toSubscriptionDTO(updatedSub), http.StatusOK, nil); err != nil {
		WriteHTTPError(w, r, processAppError(err))
	}
}
