package protocol

import (
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// Representation the Response structure in protocol:
// Status:   uint8;  1byte
// BodySize: uint32; 4bytes
// Body:     []byte; changeable size
type Response struct {
	Header ResponseHeader
	Body   []byte
}

func NewResponse(s status.Status, body []byte) *Response {
	return &Response{
		Header: ResponseHeader{
			Status:   s,
			BodySize: uint32(len(body)),
		},
		Body: body,
	}
}

func (r *Response) ToBytes() []byte {
	headerBytes := r.Header.ToBytes()

	b := make([]byte, 0, ResponseHeaderSize+len(r.Body))
	b = append(b, headerBytes[:]...)
	b = append(b, r.Body...)

	return b
}

func NewResponseFromBytes(in []byte) (*Response, error) {
	if len(in) < ResponseHeaderSize {
		return nil, status.WrongInput
	}

	var r Response
	r.Header = NewResponseHeaderFromBytes([ResponseHeaderSize]byte(in[:ResponseHeaderSize]))

	if len(in) < ResponseHeaderSize+int(r.Header.BodySize) {
		return nil, status.WrongInput
	}
	r.Body = in[ResponseHeaderSize : ResponseHeaderSize+int(r.Header.BodySize)]
	return &r, nil
}
