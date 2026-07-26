package protocol_test

import (
	"bytes"
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestRequestHeader_RoundTrip(t *testing.T) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	original := protocol.RequestHeader{
		Operation: uint8(protocol.Set),
		Key:       key,
		BodySize:  1024,
	}

	bytesData := original.ToBytes()
	parsed := protocol.NewRequestHeaderFromBytes(bytesData)

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

func TestRequest_RoundTrip(t *testing.T) {
	key := [protocol.RequestKeySize]byte{0xAA, 0xBB}
	body := []byte("hello gedis protocol")

	req := protocol.NewRequest(protocol.Set, key, body)
	if req.Header.BodySize != uint32(len(body)) {
		t.Fatalf("expected Header.BodySize=%d, got %d", len(body), req.Header.BodySize)
	}

	rawBytes := req.ToBytes()
	expectedLen := protocol.RequestHeaderSize + len(body)
	if len(rawBytes) != expectedLen {
		t.Fatalf("expected raw bytes length %d, got %d", expectedLen, len(rawBytes))
	}

	parsed, err := protocol.NewRequestFromBytes(rawBytes)
	if err != nil {
		t.Fatalf("unexpected error parsing request: %v", err)
	}

	if parsed.Header.Operation != uint8(protocol.Set) {
		t.Errorf("Operation mismatch: got %d, want %d", parsed.Header.Operation, protocol.Set)
	}
	if parsed.Header.Key != key {
		t.Errorf("Key mismatch: got %v, want %v", parsed.Header.Key, key)
	}
	if !bytes.Equal(parsed.Body, body) {
		t.Errorf("Body mismatch: got %q, want %q", string(parsed.Body), string(body))
	}
}

func TestRequest_NewFromBytes_ShortInput(t *testing.T) {
	shortData := make([]byte, protocol.RequestHeaderSize-1)
	_, err := protocol.NewRequestFromBytes(shortData)
	if err != status.WrongInput {
		t.Errorf("expected status.WrongInput for short request header, got %v", err)
	}
}

func TestResponseHeader_RoundTrip(t *testing.T) {
	original := protocol.ResponseHeader{
		Status:   status.KeyAlreadyExists,
		BodySize: 2048,
	}

	bytesData := original.ToBytes()
	parsed := protocol.NewResponseHeaderFromBytes(bytesData)

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

	rawBytes := resp.ToBytes()
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
		raw := resp.ToBytes()
		// Truncate body by 2 bytes
		truncated := raw[:len(raw)-2]

		_, err := protocol.NewResponseFromBytes(truncated)
		if err != status.WrongInput {
			t.Errorf("expected status.WrongInput for truncated body, got %v", err)
		}
	})
}
