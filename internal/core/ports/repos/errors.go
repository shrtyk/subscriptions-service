package repos

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
