package protocol

import (
	"encoding/binary"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// Representation the ResponseHeader structure in protocol:
// Status:   uint8;  1byte
// BodySize: uint32; 4bytes
type ResponseHeader struct {
	BodySize uint32
	Status   status.Status
}

func NewResponseHeaderFromBytes(bytes *[ResponseHeaderSize]byte) ResponseHeader {
	var h ResponseHeader

	h.Status = status.Status(bytes[0])
	h.BodySize = binary.BigEndian.Uint32(bytes[ResponseStatusSize : ResponseStatusSize+BodySizeSize])

	return h
}

func (h *ResponseHeader) ToBytes() *[ResponseHeaderSize]byte {
	var b [ResponseHeaderSize]byte
	h.MarshalTo(b[:])
	return &b
}

// MarshalTo serializes the header into caller-owned storage.
func (h *ResponseHeader) MarshalTo(b []byte) {
	if len(b) < ResponseHeaderSize {
		return
	}

	b[0] = byte(h.Status)
	offset := ResponseStatusSize

	binary.BigEndian.PutUint32(b[offset:offset+BodySizeSize], h.BodySize)
}
