package protocol

const (
	OperationTypeSize = 1
	KeySize           = 16
	BodySizeSize      = 3 // size of the size of the valueSize field in a header
	HeaderSize        = OperationTypeSize + KeySize + BodySizeSize
)

type Header struct {
	Key       [KeySize]byte
	Operation uint8
	BodySize  uint32
}

func NewHeaderFromBytes(in [HeaderSize]byte) (header *Header) {
	header.Operation = in[0]

	copy(header.Key[:], in[OperationTypeSize:OperationTypeSize+KeySize])

	bodySize := in[OperationTypeSize+KeySize : HeaderSize]
	header.BodySize = uint32(bodySize[0])<<16 | uint32(bodySize[1])<<8 | uint32(bodySize[2])

	return
}

func (h *Header) ToBytes() [HeaderSize]byte {
	var b [HeaderSize]byte

	b[0] = h.Operation
	offset := uint32(OperationTypeSize)

	copy(b[offset:], h.Key[:])
	offset += KeySize

	b[offset] = byte(h.BodySize >> 16)
	b[offset+1] = byte(h.BodySize >> 8)
	b[offset+2] = byte(h.BodySize)

	return b
}
