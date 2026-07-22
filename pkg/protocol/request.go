package protocol

import (
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// Representation the Request structure in protocol:
// Header:  RequestHeader;  20bytes
// Body:    []byte;         changeable size
type Request struct {
	Body   []byte
	Header *RequestHeader
}

func NewRequest(ActionType ActionType, Key [RequestKeySize]byte, Body *[]byte) *Request {
	return &Request{
		Header: &RequestHeader{
			Operation: uint8(ActionType),
			Key:       Key,
			BodySize:  uint32(len(*Body)),
		},
		Body: *Body,
	}
}

func NewRequestFromBytes(in []byte) (*Request, error) {
	var req Request
	if len(in) < RequestHeaderSize {
		return nil, status.WrongInput
	}
	req.Header = NewRequestHeaderFromBytes([RequestHeaderSize]byte(in[:RequestHeaderSize]))
	req.Body = in[RequestHeaderSize:]
	return &req, nil
}

func (r *Request) ToBytes() []byte {
	headerBytes := r.Header.ToBytes()
	b := make([]byte, RequestHeaderSize+len(r.Body))

	copy(b, headerBytes[:])
	copy(b[RequestHeaderSize:], r.Body)
	return b
}

type ActionType uint8

const (
	Set ActionType = iota
	Get
	Del
	Exist
	TTL_Set
	TTL_Get
	TTL_Del
	TTL_Exist
)
