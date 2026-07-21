package status

type Status uint32

const (
	OK Status = iota
	NO_SUCH_KEY
	SUCH_KEY_ALREADY_EXISTS
	WRONG_INPUT
	INTERNAL_ERROR
)

func (s Status) Error() string {
	switch s {
	case OK:
		return "OK"
	case NO_SUCH_KEY:
		return "No such key"
	case SUCH_KEY_ALREADY_EXISTS:
		return "Such key already exists"
	case WRONG_INPUT:
		return "Wrong input"
	case INTERNAL_ERROR:
		return "Internal error"
	default:
		return "Unknown status code"
	}
}
