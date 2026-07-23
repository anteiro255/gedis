package status

type Status uint32

const (
	OK Status = iota
	NoSuchKey
	SuchKeyAlreadyExists
	WrongInput
	InternalError
	DeadLineExceeded
)

func (s Status) Error() string {
	switch s {
	case OK:
		return "OK"
	case NoSuchKey:
		return "No such key"
	case SuchKeyAlreadyExists:
		return "Such key already exists"
	case WrongInput:
		return "Wrong input"
	case InternalError:
		return "Internal error"
	case DeadLineExceeded:
		return "Deadline exceeded"
	default:
		return "Unknown status code"
	}
}
