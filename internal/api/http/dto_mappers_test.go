package http

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/shrtyk/subscriptions-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func Test_toSubscriptionDTO(t *testing.T) {
	t.Parallel()

	endDateStr := "12-2026"
	endDate := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		sub  *domain.Subscription
		want *dto.Subscription
	}{
		{
			name: "Subscription with end date",
			sub: &domain.Subscription{
				ID:          uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
				ServiceName: "Netflix",
				MonthlyCost: 1500,
				UserID:      uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"),
				StartDate:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     &endDate,
			},
			want: &dto.Subscription{
				Id:          uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
				ServiceName: "Netflix",
				MonthlyCost: 1500,
				UserId:      uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12"),
				StartDate:   "01-2025",
				EndDate:     &endDateStr,
			},
		},
		{
			name: "Subscription without end date",
			sub: &domain.Subscription{
				ID:          uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13"),
				ServiceName: "Spotify",
				MonthlyCost: 500,
				UserID:      uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14"),
				StartDate:   time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     nil,
			},
			want: &dto.Subscription{
				Id:          uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13"),
				ServiceName: "Spotify",
				MonthlyCost: 500,
				UserId:      uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14"),
				StartDate:   "03-2024",
				EndDate:     nil,
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := toSubscriptionDTO(tc.sub)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_fromNewSubscriptionDTO(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	endDateStr := "12-2025"

	tests := []struct {
		name    string
		dto     *dto.NewSubscription
		want    *domain.Subscription
		wantErr bool
	}{
		{
			name: "Valid new subscription with end date",
			dto: &dto.NewSubscription{
				ServiceName: "New Service",
				MonthlyCost: 200,
				UserId:      userID,
				StartDate:   "01-2025",
				EndDate:     &endDateStr,
			},
			want: &domain.Subscription{
				ServiceName: "New Service",
				MonthlyCost: 200,
				UserID:      userID,
				StartDate:   startDate,
				EndDate:     &endDate,
			},
			wantErr: false,
		},
		{
			name: "Valid new subscription without end date",
			dto: &dto.NewSubscription{
				ServiceName: "Another Service",
				MonthlyCost: 100,
				UserId:      userID,
				StartDate:   "01-2025",
				EndDate:     nil,
			},
			want: &domain.Subscription{
				ServiceName: "Another Service",
				MonthlyCost: 100,
				UserID:      userID,
				StartDate:   startDate,
				EndDate:     nil,
			},
			wantErr: false,
		},
		{
			name: "Invalid start date format",
			dto: &dto.NewSubscription{
				ServiceName: "Invalid Date",
				MonthlyCost: 50,
				UserId:      userID,
				StartDate:   "2025-01",
				EndDate:     nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid end date format",
			dto: &dto.NewSubscription{
				ServiceName: "Invalid Date",
				MonthlyCost: 50,
				UserId:      userID,
				StartDate:   "01-2025",
				EndDate:     func(s string) *string { return &s }("2025-12"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Start date after end date",
			dto: &dto.NewSubscription{
				ServiceName: "Bad Dates",
				MonthlyCost: 50,
				UserId:      userID,
				StartDate:   "12-2025",
				EndDate:     func(s string) *string { return &s }("01-2025"),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := fromNewSubscriptionDTO(tc.dto)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func Test_fromUpdateSubscriptionRequest(t *testing.T) {
	t.Parallel()

	serviceName := "Updated Service"
	monthlyCost := 300
	endDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		req     *UpdateSubscriptionRequest
		want    *domain.SubscriptionUpdate
		wantErr bool
	}{
		{
			name: "Update with all fields",
			req: &UpdateSubscriptionRequest{
				ServiceName: &serviceName,
				MonthlyCost: &monthlyCost,
				EndDate:     json.RawMessage(`"01-2026"`),
			},
			want: &domain.SubscriptionUpdate{
				ServiceName: &serviceName,
				MonthlyCost: &monthlyCost,
				EndDate:     &endDate,
			},
			wantErr: false,
		},
		{
			name: "Clear end date with null",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`null`),
			},
			want: &domain.SubscriptionUpdate{
				ClearEndDate: true,
			},
			wantErr: false,
		},
		{
			name: "Absent end date",
			req: &UpdateSubscriptionRequest{
				ServiceName: &serviceName,
			},
			want: &domain.SubscriptionUpdate{
				ServiceName: &serviceName,
			},
			wantErr: false,
		},
		{
			name: "Invalid end date format",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`"2026-01"`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid end date json",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`not-a-string`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := fromUpdateSubscriptionRequest(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func Test_toListFilter(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	serviceName := "Test Service"
	page := 1
	pageSize := 10

	tests := []struct {
		name   string
		params dto.ListSubscriptionsParams
		want   domain.SubscriptionFilter
	}{
		{
			name: "All parameters provided",
			params: dto.ListSubscriptionsParams{
				UserId:      &userID,
				ServiceName: &serviceName,
				Page:        &page,
				PageSize:    &pageSize,
			},
			want: domain.SubscriptionFilter{
				UserID:      &userID,
				ServiceName: &serviceName,
				Page:        &page,
				PageSize:    &pageSize,
			},
		},
		{
			name: "Only UserID",
			params: dto.ListSubscriptionsParams{
				UserId: &userID,
			},
			want: domain.SubscriptionFilter{
				UserID: &userID,
			},
		},
		{
			name: "Only ServiceName",
			params: dto.ListSubscriptionsParams{
				ServiceName: &serviceName,
			},
			want: domain.SubscriptionFilter{
				ServiceName: &serviceName,
			},
		},
		{
			name:   "No parameters",
			params: dto.ListSubscriptionsParams{},
			want:   domain.SubscriptionFilter{},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := toListFilter(tc.params)
			assert.Equal(t, tc.want, got)
		})
	}
}
