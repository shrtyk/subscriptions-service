package subservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
	reposmocks "github.com/shrtyk/subscriptions-service/internal/core/ports/repos/mocks"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/subservice"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/tx"
	txmocks "github.com/shrtyk/subscriptions-service/internal/core/ports/tx/mocks"
	"github.com/shrtyk/subscriptions-service/pkg/errkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type serviceTestBundle struct {
	svc        subservice.SubscriptionsService
	repo       *reposmocks.MockSubscriptionRepository
	txProvider *txmocks.MockProvider
}

func setup(t *testing.T) serviceTestBundle {
	t.Helper()
	repo := reposmocks.NewMockSubscriptionRepository(t)
	txProvider := txmocks.NewMockProvider(t)
	svc := New(repo, txProvider)
	return serviceTestBundle{
		svc:        svc,
		repo:       repo,
		txProvider: txProvider,
	}
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	expectedSub := &domain.Subscription{ID: subID}
	repoErrNotFound := errkit.WrapErr("op", repos.KindNotFound, errors.New("not found"))
	genericErr := errors.New("some generic error")
	repoErrGeneric := errkit.WrapErr("op", repos.KindUnknown, genericErr)

	testCases := []struct {
		name       string
		setupMocks func(bundle serviceTestBundle)
		assertFunc func(t *testing.T, sub *domain.Subscription, err error)
	}{
		{
			name: "Success",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("GetByID", ctx, subID).Return(expectedSub, nil).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.NoError(t, err)
				assert.Equal(t, expectedSub, sub)
			},
		},
		{
			name: "Not Found",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("GetByID", ctx, subID).Return(nil, repoErrNotFound).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, sub)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindNotFound, svcErr.Kind)
			},
		},
		{
			name: "Generic Repo Error",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("GetByID", ctx, subID).Return(nil, repoErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, sub)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindUnknown, svcErr.Kind)
				assert.ErrorIs(t, err, genericErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle)
			sub, err := bundle.svc.GetByID(ctx, subID)
			tc.assertFunc(t, sub, err)
		})
	}
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	sub := domain.Subscription{ServiceName: "Test"}
	repoErrDuplicate := errkit.WrapErr("op", repos.KindDuplicate, errors.New("duplicate"))

	testCases := []struct {
		name       string
		setupMocks func(bundle serviceTestBundle)
		assertFunc func(t *testing.T, sub *domain.Subscription, err error)
	}{
		{
			name: "Success",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.NoError(t, err)
				assert.NotNil(t, sub)
			},
		},
		{
			name: "Duplicate",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(repoErrDuplicate).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, sub)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindBusinessLogic, svcErr.Kind)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle)
			createdSub, err := bundle.svc.Create(ctx, sub)
			tc.assertFunc(t, createdSub, err)
		})
	}
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	repoErrNotFound := errkit.WrapErr("op", repos.KindNotFound, errors.New("not found"))
	repoErrGeneric := errkit.WrapErr("op", repos.KindUnknown, errors.New("db is down"))

	testCases := []struct {
		name       string
		setupMocks func(bundle serviceTestBundle)
		assertFunc func(t *testing.T, err error)
	}{
		{
			name: "Success",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("Delete", ctx, subID).Return(nil).Once()
			},
			assertFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Not Found",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("Delete", ctx, subID).Return(repoErrNotFound).Once()
			},
			assertFunc: func(t *testing.T, err error) {
				require.Error(t, err)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindNotFound, svcErr.Kind)
			},
		},
		{
			name: "Generic Error",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("Delete", ctx, subID).Return(repoErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, err error) {
				require.Error(t, err)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindUnknown, svcErr.Kind)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle)
			err := bundle.svc.Delete(ctx, subID)
			tc.assertFunc(t, err)
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	repoErrNotFound := errkit.WrapErr("op", repos.KindNotFound, errors.New("not found"))
	repoErrDuplicate := errkit.WrapErr("op", repos.KindDuplicate, errors.New("duplicate value"))

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	timePtr := func(t time.Time) *time.Time { return &t }

	testCases := []struct {
		name        string
		update      domain.SubscriptionUpdate
		existingSub *domain.Subscription
		setupMocks  func(bundle serviceTestBundle, existingSub *domain.Subscription)
		assertFunc  func(t *testing.T, sub *domain.Subscription, err error)
	}{
		{
			name: "Success - Full Update",
			update: domain.SubscriptionUpdate{
				ServiceName: strPtr("New Name"),
				MonthlyCost: intPtr(200),
				EndDate:     timePtr(time.Now().Add(24 * time.Hour)),
			},
			existingSub: &domain.Subscription{ID: subID, ServiceName: "Old Name", MonthlyCost: 100},
			setupMocks: func(bundle serviceTestBundle, existingSub *domain.Subscription) {
				bundle.repo.On("GetByID", ctx, subID).Return(existingSub, nil).Once()
				bundle.repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil).Once()
				bundle.txProvider.On("WithTransaction", ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(uow tx.UnitOfWork) error) error {
						uowMock := txmocks.NewMockUnitOfWork(t)
						uowMock.On("Subscriptions").Return(bundle.repo)
						return fn(uowMock)
					}).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.NoError(t, err)
				assert.NotNil(t, sub)
				assert.Equal(t, "New Name", sub.ServiceName)
				assert.Equal(t, 200, sub.MonthlyCost)
				assert.NotNil(t, sub.EndDate)
			},
		},
		{
			name:        "Success - Clear EndDate",
			update:      domain.SubscriptionUpdate{ClearEndDate: true},
			existingSub: &domain.Subscription{ID: subID, ServiceName: "Old Name", EndDate: timePtr(time.Now())},
			setupMocks: func(bundle serviceTestBundle, existingSub *domain.Subscription) {
				bundle.repo.On("GetByID", ctx, subID).Return(existingSub, nil).Once()
				bundle.repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil).Once()
				bundle.txProvider.On("WithTransaction", ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(uow tx.UnitOfWork) error) error {
						uowMock := txmocks.NewMockUnitOfWork(t)
						uowMock.On("Subscriptions").Return(bundle.repo)
						return fn(uowMock)
					}).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.NoError(t, err)
				assert.NotNil(t, sub)
				assert.Nil(t, sub.EndDate)
			},
		},
		{
			name:        "GetByID returns Not Found",
			update:      domain.SubscriptionUpdate{ServiceName: strPtr("New Name")},
			existingSub: nil,
			setupMocks: func(bundle serviceTestBundle, existingSub *domain.Subscription) {
				bundle.repo.On("GetByID", ctx, subID).Return(nil, repoErrNotFound).Once()
				bundle.txProvider.On("WithTransaction", ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(uow tx.UnitOfWork) error) error {
						uowMock := txmocks.NewMockUnitOfWork(t)
						uowMock.On("Subscriptions").Return(bundle.repo)
						return fn(uowMock)
					}).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, sub)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindNotFound, svcErr.Kind)
			},
		},
		{
			name:        "Update returns Duplicate",
			update:      domain.SubscriptionUpdate{ServiceName: strPtr("Duplicate Name")},
			existingSub: &domain.Subscription{ID: subID, ServiceName: "Old Name"},
			setupMocks: func(bundle serviceTestBundle, existingSub *domain.Subscription) {
				bundle.repo.On("GetByID", ctx, subID).Return(existingSub, nil).Once()
				bundle.repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(repoErrDuplicate).Once()
				bundle.txProvider.On("WithTransaction", ctx, mock.Anything).
					Return(func(ctx context.Context, fn func(uow tx.UnitOfWork) error) error {
						uowMock := txmocks.NewMockUnitOfWork(t)
						uowMock.On("Subscriptions").Return(bundle.repo)
						return fn(uowMock)
					}).Once()
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, sub)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindBusinessLogic, svcErr.Kind)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle, tc.existingSub)
			updatedSub, err := bundle.svc.Update(ctx, subID, tc.update)
			tc.assertFunc(t, updatedSub, err)
		})
	}
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	filter := domain.SubscriptionFilter{}
	expectedSubs := []domain.Subscription{{ID: uuid.New()}}
	repoErrGeneric := errkit.WrapErr("op", repos.KindUnknown, errors.New("db error"))

	testCases := []struct {
		name       string
		setupMocks func(bundle serviceTestBundle)
		assertFunc func(t *testing.T, subs []domain.Subscription, err error)
	}{
		{
			name: "Success",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("List", ctx, filter).Return(expectedSubs, nil).Once()
			},
			assertFunc: func(t *testing.T, subs []domain.Subscription, err error) {
				require.NoError(t, err)
				assert.Equal(t, expectedSubs, subs)
			},
		},
		{
			name: "Generic Repo Error",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("List", ctx, filter).Return(nil, repoErrGeneric).Once()
			},
			assertFunc: func(t *testing.T, subs []domain.Subscription, err error) {
				require.Error(t, err)
				assert.Nil(t, subs)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindUnknown, svcErr.Kind)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle)
			subs, err := bundle.svc.List(ctx, filter)
			tc.assertFunc(t, subs, err)
		})
	}
}

func TestService_TotalCost(t *testing.T) {
	ctx := context.Background()
	filter := domain.SubscriptionFilter{}
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	subs := []domain.Subscription{
		{
			MonthlyCost: 100,
			StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			MonthlyCost: 50,
			StartDate:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	testCases := []struct {
		name       string
		setupMocks func(bundle serviceTestBundle)
		assertFunc func(t *testing.T, cost int, err error)
	}{
		{
			name: "Success",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("List", ctx, filter).Return(subs, nil).Once()
			},
			assertFunc: func(t *testing.T, cost int, err error) {
				require.NoError(t, err)
				assert.Equal(t, 12*100+3*50, cost)
			},
		},
		{
			name: "Repo List fails",
			setupMocks: func(bundle serviceTestBundle) {
				bundle.repo.On("List", ctx, filter).Return(nil, errors.New("db error")).Once()
			},
			assertFunc: func(t *testing.T, cost int, err error) {
				require.Error(t, err)
				assert.Zero(t, cost)
				var svcErr *errkit.BaseErr[subservice.ServiceKind]
				require.ErrorAs(t, err, &svcErr)
				assert.Equal(t, subservice.KindUnknown, svcErr.Kind)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bundle := setup(t)
			tc.setupMocks(bundle)
			cost, err := bundle.svc.TotalCost(ctx, filter, start, end)
			tc.assertFunc(t, cost, err)
		})
	}
}

func Test_calculateCostForSubscription(t *testing.T) {
	parseDate := func(s string) time.Time {
		tm, err := time.Parse("2006-01", s)
		require.NoError(t, err)
		return tm
	}
	ptrDate := func(s string) *time.Time {
		tm := parseDate(s)
		return &tm
	}

	periodStart := parseDate("2025-01")
	periodEnd := parseDate("2025-12")

	testCases := []struct {
		name         string
		sub          domain.Subscription
		expectedCost int
	}{
		{
			name:         "Full overlap, no end date",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-01")},
			expectedCost: 1200,
		},
		{
			name:         "Partial overlap, starts during period",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-06")},
			expectedCost: 700,
		},
		{
			name:         "Partial overlap, ends during period",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-01"), EndDate: ptrDate("2025-06")},
			expectedCost: 600,
		},
		{
			name:         "Subscription contained within period",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-03"), EndDate: ptrDate("2025-08")},
			expectedCost: 600,
		},
		{
			name:         "Period contained within subscription",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2024-01"), EndDate: ptrDate("2026-12")},
			expectedCost: 1200,
		},
		{
			name:         "No overlap, sub ends before period starts",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2024-01"), EndDate: ptrDate("2024-12")},
			expectedCost: 0,
		},
		{
			name:         "No overlap, sub starts after period ends",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2026-01")},
			expectedCost: 0,
		},
		{
			name:         "Single month subscription within period",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-06"), EndDate: ptrDate("2025-06")},
			expectedCost: 100,
		},
		{
			name:         "Invalid sub, starts after ends",
			sub:          domain.Subscription{MonthlyCost: 100, StartDate: parseDate("2025-06"), EndDate: ptrDate("2025-05")},
			expectedCost: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cost := calculateCostForSubscription(tc.sub, periodStart, periodEnd)
			assert.Equal(t, tc.expectedCost, cost)
		})
	}
}
