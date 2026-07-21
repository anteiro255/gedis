package protocol

type Response struct {
	Ok       bool
	BodySize uint32
	Body     []byte
}

func (r *Response) ToBytes() []byte {
	//TODO
	return nil
}

func NewResponseFromBytes(in []byte) (*Response, error) {
	//TODO
	return nil, nil
}
