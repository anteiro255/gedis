package protocol

import (
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// Representation the Request structure in protocol:
// Header:   Header;  20bytes
// Body:     []byte  changeable size
type Request struct {
	Body   []byte
	Header *Header
}

func NewRequestFromBytes(in []byte) (*Request, error) {
	var req Request
	if len(in) < HeaderSize {
		return nil, status.WrongInput
	}
	req.Header = NewHeaderFromBytes([HeaderSize]byte(in[:HeaderSize]))
	req.Body = in[HeaderSize:]
	return &req, nil
}

func (r *Request) ToBytes() []byte {
	headerBytes := r.Header.ToBytes()
	b := make([]byte, HeaderSize+len(r.Body))

	copy(b, headerBytes[:])
	copy(b[HeaderSize:], r.Body)
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
