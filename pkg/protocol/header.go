package protocol

import "encoding/binary"

const (
	OperationTypeSize = 1
	KeySize           = 16
	BodySizeSize      = 4 // size of the size of the valueSize field in a header
	HeaderSize        = OperationTypeSize + KeySize + BodySizeSize
)

// Representation the Header structure in protocol:
// Key:       [16]byte;  16bytes
// Operation: uint8;     1byte
// BodySize:  uint24;    4bytes
type Header struct {
	Key       [KeySize]byte
	BodySize  uint32
	Operation uint8
}

func NewHeaderFromBytes(in [HeaderSize]byte) *Header {
	var h Header

	h.Operation = in[0]

	copy(h.Key[:], in[OperationTypeSize:OperationTypeSize+KeySize])

	h.BodySize = binary.BigEndian.Uint32(in[OperationTypeSize+KeySize : HeaderSize])

	return &h
}

func (h *Header) ToBytes() [HeaderSize]byte {
	var b [HeaderSize]byte

	b[0] = h.Operation
	offset := uint32(OperationTypeSize)

	copy(b[offset:], h.Key[:])
	offset += KeySize

	binary.BigEndian.PutUint32(b[offset:offset+4], h.BodySize)
	offset += BodySizeSize

	return b
}
