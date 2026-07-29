package protocol_test

import (
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/action"
)

func FuzzNewRequestFromBytes(f *testing.F) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := []byte("fuzz_body")
	req := protocol.NewRequest(action.Set, key, body)
	f.Add(req.ToBytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		parsed, err := protocol.NewRequestFromBytes(data)
		if err != nil {
			return
		}
		if parsed.Header.BodySize != uint32(len(parsed.Body)) {
			t.Errorf("BodySize mismatch: header=%d, actual=%d", parsed.Header.BodySize, len(parsed.Body))
		}
	})
}

func FuzzNewResponseFromBytes(f *testing.F) {
	body := []byte("fuzz_response_body")
	resp := protocol.NewResponse(0, body)
	f.Add(resp.ToBytes())

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
