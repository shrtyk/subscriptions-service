package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shrtyk/subscriptions-service/internal/config"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/shrtyk/subscriptions-service/internal/core/ports/repos"
	"github.com/shrtyk/subscriptions-service/pkg/errkit"
)

func setup(t *testing.T) (*subsRepo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	cfg := config.RepoConfig{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}
	repo := NewSubsRepo(db, &cfg)

	t.Cleanup(func() {
		mock.ExpectClose()
		if cerr := db.Close(); cerr != nil {
			assert.NoError(t, cerr)
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	return repo, mock
}

func ptr[T any](value T) *T {
	return &value
}

func TestSubsRepo_Create(t *testing.T) {
	ctx := context.Background()

	sub := &domain.Subscription{
		ServiceName: "Test Service",
		MonthlyCost: 100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	dbErr := errors.New("generic DB error")
	generatedID := uuid.New()
	generatedTime := time.Now()

	testCases := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock, sub *domain.Subscription)
		assertFunc  func(t *testing.T, err error, sub *domain.Subscription)
		expectedErr error
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(generatedID, generatedTime, generatedTime)

				mock.ExpectQuery(createQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate).
					WillReturnRows(rows)
			},
			assertFunc: func(t *testing.T, err error, sub *domain.Subscription) {
				require.NoError(t, err)
				assert.Equal(t, generatedID, sub.ID)
				assert.Equal(t, generatedTime, sub.CreatedAt)
				assert.Equal(t, generatedTime, sub.UpdatedAt)
			},
		},
		{
			name: "Duplicate",
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				pgErr := &pgconn.PgError{Code: "23505"}
				mock.ExpectQuery(createQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate).
					WillReturnError(pgErr)
			},
			assertFunc: func(t *testing.T, err error, sub *domain.Subscription) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindDuplicate, baseErr.Kind)
			},
		},
		{
			name: "Generic DB Error",
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				mock.ExpectQuery(createQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate).
					WillReturnError(dbErr)
			},
			assertFunc: func(t *testing.T, err error, sub *domain.Subscription) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.ErrorIs(t, err, dbErr)
			},
			expectedErr: dbErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := setup(t)
			tc.setupMock(mock, sub)
			err := repo.Create(ctx, sub)
			tc.assertFunc(t, err, sub)
		})
	}
}

func TestSubsRepo_GetByID(t *testing.T) {
	ctx := context.Background()

	subID := uuid.New()
	expectedSub := &domain.Subscription{
		ID:          subID,
		ServiceName: "Test Service",
		MonthlyCost: 100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	testCases := []struct {
		name       string
		subID      uuid.UUID
		setupMock  func(mock sqlmock.Sqlmock, id uuid.UUID)
		assertFunc func(t *testing.T, sub *domain.Subscription, err error)
	}{
		{
			name:  "Success",
			subID: subID,
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				rows := sqlmock.NewRows(
					[]string{"id", "service_name", "monthly_cost", "user_id",
						"start_date", "end_date", "created_at", "updated_at"}).
					AddRow(expectedSub.ID, expectedSub.ServiceName, expectedSub.MonthlyCost,
						expectedSub.UserID, expectedSub.StartDate, expectedSub.EndDate, time.Now(), time.Now())
				mock.ExpectQuery(getByIDQuery).WithArgs(id).WillReturnRows(rows)
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedSub.ID, sub.ID)
				assert.Equal(t, expectedSub.ServiceName, sub.ServiceName)
			},
		},
		{
			name:  "Not Found",
			subID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectQuery(getByIDQuery).WithArgs(id).WillReturnError(sql.ErrNoRows)
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindNotFound, baseErr.Kind)
				assert.Nil(t, sub)
			},
		},
		{
			name:  "Scan Error",
			subID: subID,
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow("not-a-uuid")
				mock.ExpectQuery(getByIDQuery).WithArgs(id).WillReturnRows(rows)
			},
			assertFunc: func(t *testing.T, sub *domain.Subscription, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.Nil(t, sub)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := setup(t)
			tc.setupMock(mock, tc.subID)
			sub, err := repo.GetByID(ctx, tc.subID)
			tc.assertFunc(t, sub, err)
		})
	}
}

func TestSubsRepo_Update(t *testing.T) {
	ctx := context.Background()

	sub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Updated Service",
		MonthlyCost: 200,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	dbErr := errors.New("db error")
	rowsAffectedErr := errors.New("rows affected error")

	testCases := []struct {
		name        string
		sub         *domain.Subscription
		setupMock   func(mock sqlmock.Sqlmock, sub *domain.Subscription)
		assertFunc  func(t *testing.T, err error)
		expectedErr error
	}{
		{
			name: "Success",
			sub:  sub,
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				mock.ExpectExec(updateQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			assertFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Not Found",
			sub:  sub,
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				mock.ExpectExec(updateQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			assertFunc: func(t *testing.T, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindNotFound, baseErr.Kind)
			},
		},
		{
			name: "Generic DB Error",
			sub:  sub,
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				mock.ExpectExec(updateQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
					WillReturnError(dbErr)
			},
			assertFunc: func(t *testing.T, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.ErrorIs(t, err, dbErr)
			},
			expectedErr: dbErr,
		},
		{
			name: "RowsAffected Error",
			sub:  sub,
			setupMock: func(mock sqlmock.Sqlmock, sub *domain.Subscription) {
				mock.ExpectExec(updateQuery).
					WithArgs(sub.ServiceName, sub.MonthlyCost, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
					WillReturnResult(sqlmock.NewErrorResult(rowsAffectedErr))
			},
			assertFunc: func(t *testing.T, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.ErrorIs(t, err, rowsAffectedErr)
			},
			expectedErr: rowsAffectedErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := setup(t)
			tc.setupMock(mock, tc.sub)
			err := repo.Update(ctx, tc.sub)
			tc.assertFunc(t, err)
		})
	}
}

func TestSubsRepo_Delete(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()

	dbErr := errors.New("db error")

	testCases := []struct {
		name        string
		subID       uuid.UUID
		setupMock   func(mock sqlmock.Sqlmock, id uuid.UUID)
		assertFunc  func(t *testing.T, err error)
		expectedErr error
	}{
		{
			name:  "Success",
			subID: subID,
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectExec(deleteQuery).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			assertFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "Not Found",
			subID: subID,
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectExec(deleteQuery).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			assertFunc: func(t *testing.T, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindNotFound, baseErr.Kind)
			},
		},
		{
			name:  "Generic DB Error",
			subID: subID,
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectExec(deleteQuery).
					WithArgs(id).
					WillReturnError(dbErr)
			},
			assertFunc: func(t *testing.T, err error) {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.ErrorIs(t, err, dbErr)
			},
			expectedErr: dbErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := setup(t)
			tc.setupMock(mock, tc.subID)
			err := repo.Delete(ctx, tc.subID)
			tc.assertFunc(t, err)
		})
	}
}

func TestSubsRepo_List(t *testing.T) {
	repo, mock := setup(t)
	ctx := context.Background()

	userID := uuid.New()
	serviceName := "Test Service"

	mockCols := []string{"id", "service_name", "monthly_cost", "user_id", "start_date", "end_date", "created_at", "updated_at"}

	testCases := []struct {
		name         string
		filter       domain.SubscriptionFilter
		expectedSQL  string
		expectedArgs []any
		mockRows     *sqlmock.Rows
		mockErr      error
	}{
		{
			name:         "No filter, default pagination",
			filter:       domain.SubscriptionFilter{},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 10 OFFSET 0",
			expectedArgs: []any{},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "Filter by UserID",
			filter:       domain.SubscriptionFilter{UserID: ptr(userID)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE user_id = $1 LIMIT 10 OFFSET 0",
			expectedArgs: []any{userID},
			mockRows:     sqlmock.NewRows(mockCols).AddRow(uuid.New(), serviceName, 100, userID, time.Now(), nil, time.Now(), time.Now()),
		},
		{
			name:         "Filter by ServiceName",
			filter:       domain.SubscriptionFilter{ServiceName: ptr(serviceName)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE service_name = $1 LIMIT 10 OFFSET 0",
			expectedArgs: []any{serviceName},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "Filter by UserID and ServiceName",
			filter:       domain.SubscriptionFilter{UserID: ptr(userID), ServiceName: ptr(serviceName)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE user_id = $1 AND service_name = $2 LIMIT 10 OFFSET 0",
			expectedArgs: []any{userID, serviceName},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "Custom Page and PageSize",
			filter:       domain.SubscriptionFilter{Page: ptr(3), PageSize: ptr(20)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 20 OFFSET 40",
			expectedArgs: []any{},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "PageSize exceeds MaxPageSize",
			filter:       domain.SubscriptionFilter{PageSize: ptr(200)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 100 OFFSET 0",
			expectedArgs: []any{},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "Page 1 with custom PageSize",
			filter:       domain.SubscriptionFilter{Page: ptr(1), PageSize: ptr(5)},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 5 OFFSET 0",
			expectedArgs: []any{},
			mockRows:     sqlmock.NewRows(mockCols),
		},
		{
			name:         "DB Query Error",
			filter:       domain.SubscriptionFilter{},
			expectedSQL:  "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions LIMIT 10 OFFSET 0",
			expectedArgs: []any{},
			mockErr:      errors.New("db query error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			driverArgs := make([]driver.Value, len(tc.expectedArgs))
			for i, v := range tc.expectedArgs {
				driverArgs[i] = v
			}

			query := mock.ExpectQuery(tc.expectedSQL).WithArgs(driverArgs...)
			if tc.mockErr != nil {
				query.WillReturnError(tc.mockErr)
			} else {
				query.WillReturnRows(tc.mockRows)
			}

			subs, err := repo.List(ctx, tc.filter)
			if tc.mockErr != nil {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.Nil(t, subs)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSubsRepo_ListAll(t *testing.T) {
	repo, mock := setup(t)
	ctx := context.Background()

	userID := uuid.New()
	serviceName := "Test Service"

	mockCols := []string{"id", "service_name", "monthly_cost", "user_id", "start_date", "end_date", "created_at", "updated_at"}

	testCases := []struct {
		name        string
		filter      domain.SubscriptionFilter
		expectedSQL string
		mockRows    *sqlmock.Rows
		mockErr     error
	}{
		{
			name:        "Success - No filter",
			filter:      domain.SubscriptionFilter{},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions",
			mockRows:    sqlmock.NewRows(mockCols),
		},
		{
			name:        "Success - Filter by UserID",
			filter:      domain.SubscriptionFilter{UserID: ptr(userID)},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE user_id = $1",
			mockRows:    sqlmock.NewRows(mockCols),
		},
		{
			name:        "Success - Filter by ServiceName",
			filter:      domain.SubscriptionFilter{ServiceName: ptr(serviceName)},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE service_name = $1",
			mockRows:    sqlmock.NewRows(mockCols),
		},
		{
			name:        "Success - Filter by UserID and ServiceName",
			filter:      domain.SubscriptionFilter{UserID: ptr(userID), ServiceName: ptr(serviceName)},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE user_id = $1 AND service_name = $2",
			mockRows:    sqlmock.NewRows(mockCols),
		},
		{
			name:        "DB Query Error",
			filter:      domain.SubscriptionFilter{},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions",
			mockErr:     errors.New("db query error"),
		},
		{
			name:        "Scan Error",
			filter:      domain.SubscriptionFilter{},
			expectedSQL: "SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at FROM subscriptions",
			mockRows:    sqlmock.NewRows([]string{"id"}).AddRow("not-a-uuid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var expectedArgs []driver.Value
			if tc.filter.UserID != nil {
				expectedArgs = append(expectedArgs, *tc.filter.UserID)
			}
			if tc.filter.ServiceName != nil {
				expectedArgs = append(expectedArgs, *tc.filter.ServiceName)
			}

			query := mock.ExpectQuery(tc.expectedSQL).WithArgs(expectedArgs...)

			if tc.mockErr != nil {
				query.WillReturnError(tc.mockErr)
			} else {
				query.WillReturnRows(tc.mockRows)
			}

			subs, err := repo.ListAll(ctx, tc.filter)

			if tc.mockErr != nil || tc.name == "Scan Error" {
				var baseErr *errkit.BaseErr[repos.RepoKind]
				require.ErrorAs(t, err, &baseErr)
				assert.Equal(t, repos.KindUnknown, baseErr.Kind)
				assert.Nil(t, subs)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, subs)
			}
		})
	}
}
