package protocol

const (
	// General
	BodySizeSize = 4 // size of the size of the valueSize field in a header

	// Request
	RequestOperationTypeSize = 1
	RequestKeySize           = 16
	RequestHeaderSize        = RequestOperationTypeSize + RequestKeySize + BodySizeSize

	// Response
	ResponseStatusSize = 1
	ResponseHeaderSize = ResponseStatusSize + BodySizeSize
)
