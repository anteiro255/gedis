package protocol

import (
	"github.com/anteiro255/gedis/pkg/protocol/action"
)

// Representation the Request structure in protocol:
// Header:  RequestHeader;  20bytes
// Body:    []byte;         changeable size
type Request struct {
	Body   []byte
	Header *RequestHeader
}

func NewRequest(action action.Action, key [RequestKeySize]byte, body []byte) *Request {
	return &Request{
		Header: &RequestHeader{
			Operation: uint8(action),
			Key:       key,
			BodySize:  uint32(len(body)),
		},
		Body: body,
	}
}
