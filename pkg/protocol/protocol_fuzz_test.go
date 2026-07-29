package protocol_test

import (
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol"
)

func FuzzNewResponseFromBytes(f *testing.F) {
	body := []byte("fuzz_response_body")
	resp := protocol.NewResponse(0, body)
	headerBytes := resp.Header.ToBytes()
	raw := make([]byte, protocol.ResponseHeaderSize+len(resp.Body))
	copy(raw, headerBytes[:])
	copy(raw[protocol.ResponseHeaderSize:], resp.Body)
	f.Add(raw)

	f.Fuzz(func(t *testing.T, data []byte) {
		parsed, err := protocol.NewResponseFromBytes(data)
		if err != nil {
			return
		}
		if parsed.Header.BodySize != uint32(len(parsed.Body)) {
			t.Errorf("BodySize mismatch: header=%d, actual=%d", parsed.Header.BodySize, len(parsed.Body))
		}
	})
}