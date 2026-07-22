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

func NewResponseHeaderFromBytes(in [ResponseHeaderSize]byte) ResponseHeader {
	var h ResponseHeader

	h.Status = status.Status(in[0])
	h.BodySize = binary.BigEndian.Uint32(in[ResponseStatusSize : ResponseStatusSize+BodySizeSize])

	return h
}

func (h *ResponseHeader) ToBytes() [ResponseHeaderSize]byte {
	var b [ResponseHeaderSize]byte

	b[0] = byte(h.Status)
	offset := uint32(ResponseStatusSize)

	binary.BigEndian.PutUint32(b[offset:offset+BodySizeSize], h.BodySize)
	offset += BodySizeSize

	return b
}
