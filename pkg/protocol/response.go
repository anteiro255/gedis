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

func NewResponseFromBytes(resBytes []byte) (*Response, error) {
	if len(resBytes) < ResponseHeaderSize {
		return nil, status.WrongInput
	}

	var r Response
	headerBytes := [ResponseHeaderSize]byte(resBytes[:ResponseHeaderSize])
	r.Header = NewResponseHeaderFromBytes(&headerBytes)

	if len(resBytes) < ResponseHeaderSize+int(r.Header.BodySize) {
		return nil, status.WrongInput
	}
	r.Body = resBytes[ResponseHeaderSize : ResponseHeaderSize+int(r.Header.BodySize)]
	return &r, nil
}
