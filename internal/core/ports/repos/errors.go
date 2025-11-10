package repos

import "github.com/shrtyk/subscriptions-service/pkg/errkit"

type RepoKind int

const (
	KindUnknown RepoKind = iota
	KindNotFound
	KindDuplicate
	KindInvalidArgument
)

func (k RepoKind) String() string {
	switch k {
	case KindNotFound:
		return "NotFound"
	case KindDuplicate:
		return "Duplicate"
	case KindInvalidArgument:
		return "InvalidArgument"
	default:
		return "Unknown"
	}
}

func NewErr(op string, kind RepoKind) error {
	return errkit.NewErr(op, kind)
}

func WrapErr(op string, kind RepoKind, cause error) error {
	return errkit.WrapErr(op, kind, cause)
}
