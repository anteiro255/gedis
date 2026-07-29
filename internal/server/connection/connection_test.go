package connection_test

import (
	"bytes"
	"net"
	"sync"
	"testing"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// connection exposes read/write the same way ConnectionWithClient does
type connection struct {
	net.Conn
	cfg *config.Config
}

func (c *connection) read() (*protocol.Request, error) {
	var firstByte [1]byte
	_, err := c.Read(firstByte[:])
	if err != nil {
		return nil, err
	}
	var headerAsBytes [protocol.RequestHeaderSize]byte
	headerAsBytes[0] = firstByte[0]
	_, err = c.Read(headerAsBytes[1:])
	if err != nil {
		return nil, err
	}
	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)
	req.Body = make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = c.Read(req.Body)
		if err != nil {
			return nil, err
		}
	}
	return &req, nil
}

func (c *connection) writeResponse(resp *protocol.Response) error {
	respBytes := resp.ToBytes()
	totalWritten := 0
	for totalWritten < len(respBytes) {
		n, err := c.Write(respBytes[totalWritten:])
		if err != nil {
			return err
		}
		if n == 0 {
			return net.ErrClosed
		}
		totalWritten += n
	}
	return nil
}

func TestConnectionReadRequest(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	conn := &connection{Conn: serverConn, cfg: config.Default()}

	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := []byte("test_body_value")
	req := protocol.NewRequest(action.Set, key, body)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotReq, err := conn.read()
		if err != nil {
			t.Errorf("read() error: %v", err)
			return
		}
		if gotReq.Header.Operation != uint8(action.Set) {
			t.Errorf("expected operation %d, got %d", action.Set, gotReq.Header.Operation)
		}
		if gotReq.Header.Key != key {
			t.Errorf("key mismatch")
		}
		if !bytes.Equal(gotReq.Body, body) {
			t.Errorf("body mismatch: got %q, want %q", string(gotReq.Body), string(body))
		}
	}()

	_, err := clientConn.Write(req.ToBytes())
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	wg.Wait()
}

func TestConnectionWriteResponse(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	conn := &connection{Conn: serverConn, cfg: config.Default()}
	resp := protocol.NewResponse(status.OK, []byte("response_body"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := conn.writeResponse(resp)
		if err != nil {
			t.Errorf("writeResponse() error: %v", err)
		}
	}()

	buf := make([]byte, protocol.ResponseHeaderSize+len(resp.Body))
	_, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	parsed, err := protocol.NewResponseFromBytes(buf)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if parsed.Header.Status != status.OK {
		t.Errorf("expected status OK, got %v", parsed.Header.Status)
	}
	if !bytes.Equal(parsed.Body, resp.Body) {
		t.Errorf("body mismatch: got %q, want %q", string(parsed.Body), string(resp.Body))
	}
	wg.Wait()
}

func TestConnectionReadWriteRoundTrip(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	conn := &connection{Conn: serverConn, cfg: config.Default()}

	key := [protocol.RequestKeySize]byte{0xFF, 0xEE, 0xDD}
	body := []byte("roundtrip_data")
	req := protocol.NewRequest(action.Get, key, body)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotReq, err := conn.read()
		if err != nil {
			t.Errorf("read() error: %v", err)
			return
		}
		if gotReq.Header.Operation != uint8(action.Get) {
			t.Errorf("expected operation %d, got %d", action.Get, gotReq.Header.Operation)
		}
		err = conn.writeResponse(protocol.NewResponse(status.NoSuchKey, nil))
		if err != nil {
			t.Errorf("writeResponse() error: %v", err)
		}
	}()

	_, err := clientConn.Write(req.ToBytes())
	if err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	buf := make([]byte, protocol.ResponseHeaderSize)
	_, err = clientConn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	parsedResp, err := protocol.NewResponseFromBytes(buf)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if parsedResp.Header.Status != status.NoSuchKey {
		t.Errorf("expected NoSuchKey, got %v", parsedResp.Header.Status)
	}
	wg.Wait()
}

func TestConnectionReadEOF(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	clientConn.Close()

	conn := &connection{Conn: serverConn, cfg: config.Default()}
	defer serverConn.Close()

	_, err := conn.read()
	if err == nil {
		t.Error("expected error reading from closed connection, got nil")
	}
}
