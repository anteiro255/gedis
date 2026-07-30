package protocol

import (
	"encoding/binary"
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
	h := RequestHeaderFromBytes(in)
	return &h
}

func RequestHeaderFromBytes(in *[RequestHeaderSize]byte) RequestHeader {
	var h RequestHeader

	h.Operation = in[0]

	copy(h.Key[:], in[RequestOperationTypeSize:RequestOperationTypeSize+RequestKeySize])

	h.BodySize = binary.BigEndian.Uint32(in[RequestOperationTypeSize+RequestKeySize : RequestHeaderSize])

	return h
}

func (h *RequestHeader) ToBytes() *[RequestHeaderSize]byte {
	var b [RequestHeaderSize]byte
	h.MarshalTo(b[:])
	return &b
}

// MarshalTo serializes the header into caller-owned storage. Keeping the
// buffer outside this package avoids an allocation in the socket hot path.
func (h *RequestHeader) MarshalTo(b []byte) {
	if len(b) < RequestHeaderSize {
		return
	}

	b[0] = h.Operation
	offset := RequestOperationTypeSize

	copy(b[offset:], h.Key[:])
	offset += RequestKeySize

	binary.BigEndian.PutUint32(b[offset:offset+BodySizeSize], h.BodySize)
}
