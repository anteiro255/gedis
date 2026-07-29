package protocol_test

import (
	"bytes"
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestRequestHeader_RoundTrip(t *testing.T) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	original := protocol.RequestHeader{
		Operation: uint8(action.Set),
		Key:       key,
		BodySize:  1024,
	}

	bytesData := original.ToBytes()
	parsed := protocol.NewRequestHeaderFromBytes(&bytesData)

	if parsed.Operation != original.Operation {
		t.Errorf("Operation mismatch: got %d, want %d", parsed.Operation, original.Operation)
	}
	if parsed.Key != original.Key {
		t.Errorf("Key mismatch: got %v, want %v", parsed.Key, original.Key)
	}
	if parsed.BodySize != original.BodySize {
		t.Errorf("BodySize mismatch: got %d, want %d", parsed.BodySize, original.BodySize)
	}
}

func TestResponseHeader_RoundTrip(t *testing.T) {
	original := protocol.ResponseHeader{
		Status:   status.KeyAlreadyExists,
		BodySize: 2048,
	}

	bytesData := original.ToBytes()
	parsed := protocol.NewResponseHeaderFromBytes(&bytesData)

	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: got %v, want %v", parsed.Status, original.Status)
	}
	if parsed.BodySize != original.BodySize {
		t.Errorf("BodySize mismatch: got %d, want %d", parsed.BodySize, original.BodySize)
	}
}

func TestResponse_RoundTrip(t *testing.T) {
	body := []byte("response payload")
	resp := protocol.NewResponse(status.OK, body)

	headerBytes := resp.Header.ToBytes()
	rawBytes := make([]byte, protocol.ResponseHeaderSize+len(resp.Body))
	copy(rawBytes, headerBytes[:])
	copy(rawBytes[protocol.ResponseHeaderSize:], resp.Body)
	expectedLen := protocol.ResponseHeaderSize + len(body)
	if len(rawBytes) != expectedLen {
		t.Fatalf("expected raw bytes length %d, got %d", expectedLen, len(rawBytes))
	}

	parsed, err := protocol.NewResponseFromBytes(rawBytes)
	if err != nil {
		t.Fatalf("unexpected error parsing response: %v", err)
	}

	if parsed.Header.Status != status.OK {
		t.Errorf("Status mismatch: got %v, want %v", parsed.Header.Status, status.OK)
	}
	if !bytes.Equal(parsed.Body, body) {
		t.Errorf("Body mismatch: got %q, want %q", string(parsed.Body), string(body))
	}
}

func TestResponse_NewFromBytes_Errors(t *testing.T) {
	t.Run("short header", func(t *testing.T) {
		shortData := make([]byte, protocol.ResponseHeaderSize-1)
		_, err := protocol.NewResponseFromBytes(shortData)
		if err != status.WrongInput {
			t.Errorf("expected status.WrongInput for short header, got %v", err)
		}
	})

	t.Run("truncated body", func(t *testing.T) {
		resp := protocol.NewResponse(status.OK, []byte("long body content"))
		headerBytes := resp.Header.ToBytes()
		raw := make([]byte, protocol.ResponseHeaderSize+len(resp.Body))
		copy(raw, headerBytes[:])
		copy(raw[protocol.ResponseHeaderSize:], resp.Body)
		truncated := raw[:len(raw)-2]

		_, err := protocol.NewResponseFromBytes(truncated)
		if err != status.WrongInput {
			t.Errorf("expected status.WrongInput for truncated body, got %v", err)
		}
	})
}