package protocol

import (
	"github.com/anteiro255/gedis/pkg/protocol/status"

	"encoding/binary"
)

const (
	StatusSize = 1
	// BodySizeSize = 3 // already declared in header.go
	ResponseMinSize = StatusSize + BodySizeSize
)

// Representation the Response structure in protocol:
// Status:   uint8;  1byte
// BodySize: uint32; 4bytes
// Body:     []byte; changeable size
type Response struct {
	Body     []byte
	BodySize uint32
	Status   status.Status
}

func NewResponse(s status.Status) *Response {
	return &Response{
		Status: s,
	}
}

func (r *Response) SetBody(body []byte) {
	r.BodySize = uint32(len(body))
	r.Body = body
}

func (r *Response) ToBytes() []byte {
	StatusByte := byte(r.Status)

	var BodySizeBytes [4]byte
	binary.BigEndian.PutUint32(BodySizeBytes[:], r.BodySize)

	b := make([]byte, 0, StatusSize+BodySizeSize+len(r.Body))
	b = append(b, StatusByte)
	b = append(b, BodySizeBytes[:]...)
	b = append(b, r.Body...)
	return b
}

func NewResponseFromBytes(in []byte) (*Response, error) {
	if len(in) < ResponseMinSize {
		return nil, status.WrongInput
	}

	var r Response
	r.Status = status.Status(in[0])
	r.BodySize = binary.BigEndian.Uint32(in[StatusSize : StatusSize+BodySizeSize])

	if len(in) < ResponseMinSize+int(r.BodySize) {
		return nil, status.WrongInput
	}
	r.Body = in[ResponseMinSize : ResponseMinSize+int(r.BodySize)]
	return &r, nil
}
