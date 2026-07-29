package protocol

import (
	"encoding/binary"
	"unsafe"

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

func BytesAsResponseHeader(bytes *[ResponseHeaderSize]byte) *ResponseHeader {
	return (*ResponseHeader)(unsafe.Pointer(&bytes[0]))
}

func (h *ResponseHeader) ToBytes() [ResponseHeaderSize]byte {
	var b [ResponseHeaderSize]byte

	b[0] = byte(h.Status)
	offset := uint32(ResponseStatusSize)

	binary.BigEndian.PutUint32(b[offset:offset+BodySizeSize], h.BodySize)
	offset += BodySizeSize

	return b
}

func (h *ResponseHeader) Asbytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(h)), uintptr(ResponseHeaderSize))
}
