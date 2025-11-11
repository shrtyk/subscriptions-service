package subservice

import "github.com/shrtyk/subscriptions-service/pkg/errkit"

type ServiceKind int

const (
	KindUnknown ServiceKind = iota
	KindValidation
	KindBusinessLogic
	KindNotFound
)

func (k ServiceKind) String() string {
	switch k {
	case KindValidation:
		return "Validation"
	case KindBusinessLogic:
		return "BusinessLogic"
	case KindNotFound:
		return "NotFound"
	default:
		return "Unknown"
	}
}

func NewErr(op string, kind ServiceKind) error {
	return errkit.NewErr(op, kind)
}

func WrapErr(op string, kind ServiceKind, cause error) error {
	return errkit.WrapErr(op, kind, cause)
}
