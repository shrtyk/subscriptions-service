package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateLayout_parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dateStr   string
		wantTime  time.Time
		wantErr   bool
		errString string
	}{
		{
			name:     "Valid date",
			dateStr:  "11-2025",
			wantTime: time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:      "Invalid month",
			dateStr:   "13-2025",
			wantTime:  time.Time{},
			wantErr:   true,
			errString: "parsing time \"13-2025\": month out of range",
		},
		{
			name:      "Invalid format",
			dateStr:   "2025-11",
			wantTime:  time.Time{},
			wantErr:   true,
			errString: "parsing time \"2025-11\": month out of range",
		},
		{
			name:      "Empty string",
			dateStr:   "",
			wantTime:  time.Time{},
			wantErr:   true,
			errString: "parsing time \"\" as \"01-2006\": cannot parse \"\" as \"01\"",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dateLayout.parse(tc.dateStr)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.errString, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTime, got)
			}
		})
	}
}

func TestDateLayout_format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		date     time.Time
		wantDate string
	}{
		{
			name:     "Valid date",
			date:     time.Date(2025, time.November, 15, 10, 30, 0, 0, time.UTC),
			wantDate: "11-2025",
		},
		{
			name:     "Another valid date",
			date:     time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantDate: "01-2023",
		},
		{
			name:     "Zero time",
			date:     time.Time{},
			wantDate: "01-0001", // Default format for zero time
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dateLayout.format(tc.date)
			assert.Equal(t, tc.wantDate, got)
		})
	}
}
