package status

import "strconv"

type Status uint32

const (
	OK Status = iota
	KeyAlreadyExists
	NoSuchKey
	NoTTL
	WrongInput
	InternalError
	DeadlineExceeded
	// NotLeader means a mutation was sent to a Raft follower. The response body
	// may contain the current leader address.
	NotLeader
)

func (s Status) Error() string {
	switch s {
	case OK:
		return "OK"
	case KeyAlreadyExists:
		return "Such key already exists"
	case NoSuchKey:
		return "No such key"
	case NoTTL:
		return "The key doesn't have TTL"
	case WrongInput:
		return "Wrong input"
	case InternalError:
		return "Internal error"
	case DeadlineExceeded:
		return "Deadline exceeded"
	case NotLeader:
		return "Not the Raft leader"
	default:
		return "Unknown status code: " + strconv.Itoa(int(s))
	}
}
