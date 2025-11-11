package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	MonthlyCost int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubscriptionUpdate struct {
	ServiceName  *string
	MonthlyCost  *int
	EndDate      *time.Time
	ClearEndDate bool // Flag to determine meaning of EndDate nil value (could mean 'delete' or 'do not update')
}
