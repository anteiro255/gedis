package protocol

import (
	"encoding/binary"
	"unsafe"
)

// Representation the RequestHeader structure in protocol:
// Key:       [16]byte;  16bytes
// Operation: uint8;     1byte
// BodySize:  uint24;    4bytes
type RequestHeader struct {
	Key       [RequestKeySize]byte
	BodySize  uint32
	Operation uint8
}

func NewRequestHeaderFromBytes(in *[RequestHeaderSize]byte) *RequestHeader {
	var h RequestHeader

	h.Operation = in[0]

	copy(h.Key[:], in[RequestOperationTypeSize:RequestOperationTypeSize+RequestKeySize])

	h.BodySize = binary.BigEndian.Uint32(in[RequestOperationTypeSize+RequestKeySize : RequestHeaderSize])

	return &h
}

func BytesAsHeader(in *[RequestHeaderSize]byte) *RequestHeader {
	return (*RequestHeader)(unsafe.Pointer(&in[0]))
}

func (h *RequestHeader) ToBytes() [RequestHeaderSize]byte {
	var b [RequestHeaderSize]byte

	b[0] = h.Operation
	offset := uint32(RequestOperationTypeSize)

	copy(b[offset:], h.Key[:])
	offset += RequestKeySize

	binary.BigEndian.PutUint32(b[offset:offset+BodySizeSize], h.BodySize)
	offset += BodySizeSize

	return b
}

func (h *RequestHeader) Asbytes(in *[RequestHeaderSize]byte) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(in)), uintptr(RequestHeaderSize))
}
