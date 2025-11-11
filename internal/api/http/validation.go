package http

import (
	"time"

	"github.com/shrtyk/subscriptions-service/internal/api/http/dto"
)

var (
	beginningOfTime = time.Time{}
	endOfTime       = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
)

const (
	startDateAfterEndDateMsg = "start_date cannot be after end_date"
	invalidDateMsg           = "invalid date format, expected MM-YYYY"
)

func validateGetTotalCostParams(params dto.GetTotalCostParams) (time.Time, time.Time, *DTOValidationError) {
	var start, end time.Time
	var err error

	if params.Start == nil && params.End == nil {
		return beginningOfTime, endOfTime, nil
	}

	if params.Start == nil {
		start = beginningOfTime
		end, err = dateLayout.parse(*params.End)
		if err != nil {
			return time.Time{}, time.Time{}, &DTOValidationError{ClientMessage: invalidDateMsg, InternalError: err}
		}
		return start, end, nil
	}

	if params.End == nil {
		end = endOfTime
		start, err = dateLayout.parse(*params.Start)
		if err != nil {
			return time.Time{}, time.Time{}, &DTOValidationError{ClientMessage: invalidDateMsg, InternalError: err}
		}
		return start, end, nil
	}

	start, err = dateLayout.parse(*params.Start)
	if err != nil {
		return time.Time{}, time.Time{}, &DTOValidationError{ClientMessage: invalidDateMsg, InternalError: err}
	}
	end, err = dateLayout.parse(*params.End)
	if err != nil {
		return time.Time{}, time.Time{}, &DTOValidationError{ClientMessage: invalidDateMsg, InternalError: err}
	}

	return start, end, nil
}

func validateNewSubscription(d *dto.NewSubscription) (time.Time, *time.Time, *DTOValidationError) {
	startDate, err := dateLayout.parse(d.StartDate)
	if err != nil {
		return time.Time{}, nil, &DTOValidationError{
			ClientMessage: invalidDateMsg,
			InternalError: err,
		}
	}

	var endDate *time.Time
	if d.EndDate != nil && *d.EndDate != "" {
		t, err := dateLayout.parse(*d.EndDate)
		if err != nil {
			return time.Time{}, nil, &DTOValidationError{
				ClientMessage: invalidDateMsg,
				InternalError: err,
			}
		}
		endDate = &t
	}

	if endDate != nil && startDate.After(*endDate) {
		return time.Time{}, nil, &DTOValidationError{ClientMessage: startDateAfterEndDateMsg}
	}

	return startDate, endDate, nil
}

func validateUpdateSubscription(d *dto.UpdateSubscription) (*time.Time, *DTOValidationError) {
	if d.EndDate == nil {
		return nil, nil
	}

	t, err := dateLayout.parse(*d.EndDate)
	if err != nil {
		return nil, &DTOValidationError{
			ClientMessage: invalidDateMsg,
			InternalError: err,
		}
	}
	return &t, nil
}
