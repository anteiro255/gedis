package protocol

import (
	"io"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type Request struct {
	Header *Header
	Body   []byte
}

func NewRequestFromBytes(in []byte) (*Request, error) {
	var req Request
	if len(in) < HeaderSize {
		return nil, status.WRONG_INPUT
	}
	req.Header = NewHeaderFromBytes([HeaderSize]byte(in[:HeaderSize]))
	req.Body = in[HeaderSize:]
	return &req, nil
}

func (r *Request) Write(w io.Writer) error {
	headerBytes := r.Header.ToBytes()

	_, err := w.Write(headerBytes[:])
	w.Write(r.Body)
	return err
}
