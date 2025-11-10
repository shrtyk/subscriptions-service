package domain

import "github.com/google/uuid"

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Page        *int
	PageSize    *int
}
