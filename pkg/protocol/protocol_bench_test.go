package protocol_test

import (
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func BenchmarkRequestHeaderToBytes(b *testing.B) {
	h := &protocol.RequestHeader{
		Operation: uint8(action.Set),
		Key:       [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		BodySize:  1024,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		h.ToBytes()
	}
}

func BenchmarkRequestHeaderFromBytes(b *testing.B) {
	h := &protocol.RequestHeader{
		Operation: uint8(action.Set),
		Key:       [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		BodySize:  1024,
	}
	bytesData := h.ToBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewRequestHeaderFromBytes(bytesData)
	}
}

func BenchmarkRequestToBytes(b *testing.B) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := make([]byte, 1024)
	req := protocol.NewRequest(action.Set, key, body)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		req.ToBytes()
	}
}

func BenchmarkRequestFromBytes(b *testing.B) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := make([]byte, 1024)
	req := protocol.NewRequest(action.Set, key, body)
	raw := req.ToBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewRequestFromBytes(raw)
	}
}

func BenchmarkResponseHeaderToBytes(b *testing.B) {
	h := protocol.ResponseHeader{
		Status:   status.OK,
		BodySize: 1024,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		h.ToBytes()
	}
}

func BenchmarkResponseHeaderFromBytes(b *testing.B) {
	h := protocol.ResponseHeader{
		Status:   status.OK,
		BodySize: 1024,
	}
	bytesData := h.ToBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewResponseHeaderFromBytes(bytesData)
	}
}

func BenchmarkResponseToBytes(b *testing.B) {
	body := make([]byte, 1024)
	resp := protocol.NewResponse(status.OK, body)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		resp.ToBytes()
	}
}

func BenchmarkResponseFromBytes(b *testing.B) {
	body := make([]byte, 1024)
	resp := protocol.NewResponse(status.OK, body)
	raw := resp.ToBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewResponseFromBytes(raw)
	}
}

func BenchmarkNewRequest(b *testing.B) {
	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := make([]byte, 1024)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewRequest(action.Set, key, body)
	}
}

func BenchmarkNewResponse(b *testing.B) {
	body := make([]byte, 1024)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		protocol.NewResponse(status.OK, body)
	}
}
