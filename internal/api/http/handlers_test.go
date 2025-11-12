package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
	ssmocks "github.com/shrtyk/subscriptions-service/internal/core/ports/subservice/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testHarness struct {
	h       *handler
	service *ssmocks.MockSubscriptionsService
}

func setup(t *testing.T) testHarness {
	t.Helper()
	service := ssmocks.NewMockSubscriptionsService(t)
	h := NewHandler(service)
	return testHarness{
		h:       h,
		service: service,
	}
}

func mustMarshal(t *testing.T, v any) io.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func newRequestWithChiCtx(t *testing.T, method, path string, body io.Reader, pathParams map[string]string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if pathParams != nil {
		rctx := chi.NewRouteContext()
		for k, v := range pathParams {
			rctx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	return req
}

func TestHandler_GetSubscriptionById(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	expectedSub := &domain.Subscription{
		ID:          subID,
		ServiceName: "Test",
		MonthlyCost: 100,
		UserID:      uuid.New(),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	serviceErrNotFound := subservice.NewErr("subservice.GetByID", subservice.KindNotFound)
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.GetByID", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		subID      string
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:  "Success",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("GetByID", ctx, subID).Return(expectedSub, nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var respBody dto.Subscription
				err := json.NewDecoder(rr.Body).Decode(&respBody)
				require.NoError(t, err)
				assert.Equal(t, rr.Code, http.StatusOK)
				assert.Equal(t, toSubscriptionDTO(expectedSub), &respBody)
			},
		},
		{
			name:  "Not Found",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("GetByID", ctx, subID).Return(nil, serviceErrNotFound).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusNotFound), errBody.Code)
				assert.Equal(t, "The requested resource was not found", errBody.Message)
			},
		},
		{
			name:  "Generic Service Error",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("GetByID", ctx, subID).Return(nil, serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
				assert.Equal(t, "Internal error", errBody.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := newRequestWithChiCtx(t, http.MethodGet, "/subscriptions/"+tc.subID, nil, map[string]string{"id": tc.subID})
			rr := httptest.NewRecorder()

			id, err := uuid.Parse(tc.subID)
			require.NoError(t, err)
			bundle.h.GetSubscriptionById(rr, req.WithContext(ctx), id)
			tc.assertFunc(t, rr)
		})
	}
}

func TestHandler_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	newSubDTO := dto.NewSubscription{
		ServiceName: "Test Service",
		MonthlyCost: 100,
		UserId:      userID,
		StartDate:   "01-2025",
	}
	domainSub, err := fromNewSubscriptionDTO(&newSubDTO)
	require.NoError(t, err)

	createdSub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: newSubDTO.ServiceName,
		MonthlyCost: newSubDTO.MonthlyCost,
		UserID:      userID,
		StartDate:   domainSub.StartDate,
	}

	serviceErrBizLogic := subservice.NewErr("subservice.Create", subservice.KindBusinessLogic)
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.Create", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		body       io.Reader
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			body: mustMarshal(t, newSubDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Create", ctx, *domainSub).Return(createdSub, nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var respBody dto.Subscription
				err := json.NewDecoder(rr.Body).Decode(&respBody)
				require.NoError(t, err)
				assert.Equal(t, rr.Code, http.StatusCreated)
				assert.Equal(t, toSubscriptionDTO(createdSub), &respBody)
			},
		},
		{
			name:       "Bad Request - Invalid JSON",
			body:       strings.NewReader("{invalid json"),
			setupMocks: nil,
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusBadRequest), errBody.Code)
			},
		},
		{
			name: "Validation Error - Invalid Date",
			body: mustMarshal(t, dto.NewSubscription{
				ServiceName: "Test",
				MonthlyCost: 100,
				UserId:      userID,
				StartDate:   "bad-date",
			}),
			setupMocks: nil,
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusUnprocessableEntity), errBody.Code)
				assert.Contains(t, errBody.Message, "invalid date format")
			},
		},
		{
			name: "Service Error - Business Logic",
			body: mustMarshal(t, newSubDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Create", ctx, *domainSub).Return(nil, serviceErrBizLogic).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusUnprocessableEntity), errBody.Code)
			},
		},
		{
			name: "Service Error - Generic",
			body: mustMarshal(t, newSubDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Create", ctx, *domainSub).Return(nil, serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := httptest.NewRequest(http.MethodPost, "/subscriptions", tc.body)
			rr := httptest.NewRecorder()

			bundle.h.CreateSubscription(rr, req.WithContext(ctx))

			tc.assertFunc(t, rr)
		})
	}
}

func TestHandler_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	updateDTO := dto.UpdateSubscription{
		ServiceName: func(s string) *string { return &s }("New Name"),
	}
	domainUpdate, err := fromUpdateSubscriptionDTO(&updateDTO)
	require.NoError(t, err)

	updatedSub := &domain.Subscription{
		ID:          subID,
		ServiceName: *updateDTO.ServiceName,
	}

	serviceErrNotFound := subservice.NewErr("subservice.Update", subservice.KindNotFound)
	serviceErrBizLogic := subservice.NewErr("subservice.Update", subservice.KindBusinessLogic)
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.Update", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		subID      string
		body       io.Reader
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:  "Success",
			subID: subID.String(),
			body:  mustMarshal(t, updateDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Update", ctx, subID, *domainUpdate).Return(updatedSub, nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var respBody dto.Subscription
				err := json.NewDecoder(rr.Body).Decode(&respBody)
				require.NoError(t, err)
				assert.Equal(t, rr.Code, http.StatusOK)
				assert.Equal(t, toSubscriptionDTO(updatedSub), &respBody)
			},
		},
		{
			name:       "Bad Request - Invalid JSON",
			subID:      subID.String(),
			body:       strings.NewReader("{invalid json"),
			setupMocks: nil,
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusBadRequest), errBody.Code)
			},
		},
		{
			name:  "Service Error - Not Found",
			subID: subID.String(),
			body:  mustMarshal(t, updateDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Update", ctx, subID, *domainUpdate).Return(nil, serviceErrNotFound).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusNotFound), errBody.Code)
			},
		},
		{
			name:  "Service Error - Business Logic",
			subID: subID.String(),
			body:  mustMarshal(t, updateDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Update", ctx, subID, *domainUpdate).Return(nil, serviceErrBizLogic).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusUnprocessableEntity), errBody.Code)
			},
		},
		{
			name:  "Service Error - Generic",
			subID: subID.String(),
			body:  mustMarshal(t, updateDTO),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Update", ctx, subID, *domainUpdate).Return(nil, serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := newRequestWithChiCtx(t, http.MethodPut, "/subscriptions/"+tc.subID, tc.body, map[string]string{"id": tc.subID})
			rr := httptest.NewRecorder()

			id, err := uuid.Parse(tc.subID)
			require.NoError(t, err)

			bundle.h.UpdateSubscription(rr, req.WithContext(ctx), id)
			tc.assertFunc(t, rr)
		})
	}
}

func TestHandler_DeleteSubscription(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	serviceErrNotFound := subservice.NewErr("subservice.Delete", subservice.KindNotFound)
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.Delete", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		subID      string
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:  "Success",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Delete", ctx, subID).Return(nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Code, http.StatusNoContent)
				assert.Equal(t, 0, rr.Body.Len())
			},
		},
		{
			name:  "Not Found",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Delete", ctx, subID).Return(serviceErrNotFound).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusNotFound), errBody.Code)
			},
		},
		{
			name:  "Generic Service Error",
			subID: subID.String(),
			setupMocks: func(bundle testHarness) {
				bundle.service.On("Delete", ctx, subID).Return(serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := newRequestWithChiCtx(t, http.MethodDelete, "/subscriptions/"+tc.subID, nil, map[string]string{"id": tc.subID})
			rr := httptest.NewRecorder()

			id, err := uuid.Parse(tc.subID)
			require.NoError(t, err)

			bundle.h.DeleteSubscription(rr, req.WithContext(ctx), id)
			tc.assertFunc(t, rr)
		})
	}
}

func TestHandler_ListSubscriptions(t *testing.T) {
	ctx := context.Background()
	expectedSubs := []domain.Subscription{
		{ID: uuid.New(), ServiceName: "Test 1"},
		{ID: uuid.New(), ServiceName: "Test 2"},
	}
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.List", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		params     dto.ListSubscriptionsParams
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "Success - No Filter",
			params: dto.ListSubscriptionsParams{},
			setupMocks: func(bundle testHarness) {
				bundle.service.On("List", ctx, mock.Anything).Return(expectedSubs, nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var respBody []dto.Subscription
				err := json.NewDecoder(rr.Body).Decode(&respBody)
				require.NoError(t, err)
				require.Len(t, respBody, 2)
				assert.Equal(t, rr.Code, http.StatusOK)
				assert.Equal(t, expectedSubs[0].ID, respBody[0].Id)
			},
		},
		{
			name:   "Service Error",
			params: dto.ListSubscriptionsParams{},
			setupMocks: func(bundle testHarness) {
				bundle.service.On("List", ctx, mock.Anything).Return(nil, serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
			rr := httptest.NewRecorder()

			bundle.h.ListSubscriptions(rr, req.WithContext(ctx), tc.params)
			tc.assertFunc(t, rr)
		})
	}
}

func TestHandler_GetTotalCost(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	genericErr := errors.New("generic error")
	serviceErrGeneric := subservice.WrapErr("subservice.TotalCost", subservice.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		params     dto.GetTotalCostParams
		setupMocks func(bundle testHarness)
		assertFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			params: dto.GetTotalCostParams{
				UserId: userID,
			},
			setupMocks: func(bundle testHarness) {
				bundle.service.On("TotalCost", ctx, mock.Anything, mock.Anything, mock.Anything).Return(12345, nil).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var respBody dto.TotalCost
				err := json.NewDecoder(rr.Body).Decode(&respBody)
				require.NoError(t, err)
				require.NotNil(t, respBody.TotalCost)
				assert.Equal(t, rr.Code, http.StatusOK)
				assert.Equal(t, 12345, *respBody.TotalCost)
			},
		},
		{
			name: "Validation Error - Invalid Start Date",
			params: dto.GetTotalCostParams{
				UserId: userID,
				Start:  func(s string) *string { return &s }("bad-date"),
			},
			setupMocks: nil,
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusUnprocessableEntity), errBody.Code)
				assert.Contains(t, errBody.Message, "invalid date format")
			},
		},
		{
			name: "Service Error",
			params: dto.GetTotalCostParams{
				UserId: userID,
			},
			setupMocks: func(bundle testHarness) {
				bundle.service.On("TotalCost", ctx, mock.Anything, mock.Anything, mock.Anything).Return(0, serviceErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var errBody dto.Error
				err := json.NewDecoder(rr.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, int32(http.StatusInternalServerError), errBody.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			if tc.setupMocks != nil {
				tc.setupMocks(bundle)
			}

			req := httptest.NewRequest(http.MethodGet, "/subscriptions/total_cost", nil)
			rr := httptest.NewRecorder()

			bundle.h.GetTotalCost(rr, req.WithContext(ctx), tc.params)
			tc.assertFunc(t, rr)
		})
	}
}
