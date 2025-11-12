package http

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
	"github.com/stretchr/testify/assert"
)

func Test_validateUpdateSubscriptionRequest(t *testing.T) {
	t.Parallel()

	endDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		req              *UpdateSubscriptionRequest
		wantEndDate      *time.Time
		wantClearEndDate bool
		wantErr          bool
	}{
		{
			name:             "Absent end date",
			req:              &UpdateSubscriptionRequest{},
			wantEndDate:      nil,
			wantClearEndDate: false,
			wantErr:          false,
		},
		{
			name: "Clear end date with null",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`null`),
			},
			wantEndDate:      nil,
			wantClearEndDate: true,
			wantErr:          false,
		},
		{
			name: "Valid end date",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`"01-2026"`),
			},
			wantEndDate:      &endDate,
			wantClearEndDate: false,
			wantErr:          false,
		},
		{
			name: "Invalid end date format",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`"2026-01"`),
			},
			wantErr: true,
		},
		{
			name: "Invalid end date json",
			req: &UpdateSubscriptionRequest{
				EndDate: json.RawMessage(`not-a-string`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotDate, gotClear, err := validateUpdateSubscriptionRequest(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantEndDate, gotDate)
				assert.Equal(t, tc.wantClearEndDate, gotClear)
			}
		})
	}
}

func Test_validateGetTotalCostParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  dto.GetTotalCostParams
		wantS   time.Time
		wantE   time.Time
		wantErr bool
	}{
		{
			name:    "No start or end date",
			params:  dto.GetTotalCostParams{},
			wantS:   beginningOfTime,
			wantE:   endOfTime,
			wantErr: false,
		},
		{
			name: "Only start date",
			params: dto.GetTotalCostParams{
				Start: func(s string) *string { return &s }("01-2025"),
			},
			wantS:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantE:   endOfTime,
			wantErr: false,
		},
		{
			name: "Only end date",
			params: dto.GetTotalCostParams{
				End: func(s string) *string { return &s }("12-2025"),
			},
			wantS:   beginningOfTime,
			wantE:   time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "Both start and end date",
			params: dto.GetTotalCostParams{
				Start: func(s string) *string { return &s }("01-2025"),
				End:   func(s string) *string { return &s }("12-2025"),
			},
			wantS:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantE:   time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "Invalid start date format",
			params: dto.GetTotalCostParams{
				Start: func(s string) *string { return &s }("2025-01"),
				End:   func(s string) *string { return &s }("12-2025"),
			},
			wantS:   time.Time{},
			wantE:   time.Time{},
			wantErr: true,
		},
		{
			name: "Invalid end date format",
			params: dto.GetTotalCostParams{
				Start: func(s string) *string { return &s }("01-2025"),
				End:   func(s string) *string { return &s }("2025-12"),
			},
			wantS:   time.Time{},
			wantE:   time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotS, gotE, err := validateGetTotalCostParams(tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				var validationError *DTOValidationError
				assert.ErrorAs(t, err, &validationError)
				assert.Equal(t, invalidDateMsg, validationError.ClientMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantS, gotS)
				assert.Equal(t, tc.wantE, gotE)
			}
		})
	}
}

func Test_validateNewSubscription(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	endDateStr := "12-2025"

	tests := []struct {
		name    string
		dto     *dto.NewSubscription
		wantS   time.Time
		wantE   *time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid subscription with end date",
			dto: &dto.NewSubscription{
				StartDate: "01-2025",
				EndDate:   &endDateStr,
			},
			wantS:   startDate,
			wantE:   &endDate,
			wantErr: false,
		},
		{
			name: "Valid subscription without end date",
			dto: &dto.NewSubscription{
				StartDate: "01-2025",
				EndDate:   nil,
			},
			wantS:   startDate,
			wantE:   nil,
			wantErr: false,
		},
		{
			name: "Invalid start date format",
			dto: &dto.NewSubscription{
				StartDate: "2025-01",
				EndDate:   nil,
			},
			wantS:   time.Time{},
			wantE:   nil,
			wantErr: true,
			errMsg:  invalidDateMsg,
		},
		{
			name: "Invalid end date format",
			dto: &dto.NewSubscription{
				StartDate: "01-2025",
				EndDate:   func(s string) *string { return &s }("2025-12"),
			},
			wantS:   time.Time{},
			wantE:   nil,
			wantErr: true,
			errMsg:  invalidDateMsg,
		},
		{
			name: "Start date after end date",
			dto: &dto.NewSubscription{
				StartDate: "12-2025",
				EndDate:   func(s string) *string { return &s }("01-2025"),
			},
			wantS:   time.Time{},
			wantE:   nil,
			wantErr: true,
			errMsg:  startDateAfterEndDateMsg,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotS, gotE, err := validateNewSubscription(tc.dto)
			if tc.wantErr {
				assert.Error(t, err)
				var validationError *DTOValidationError
				assert.ErrorAs(t, err, &validationError)
				assert.Equal(t, tc.errMsg, validationError.ClientMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantS, gotS)
				assert.Equal(t, tc.wantE, gotE)
			}
		})
	}
}
