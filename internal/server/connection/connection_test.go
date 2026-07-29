package connection_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/server/connection"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type mockConn struct {
	net.Conn
	cfg *config.Config
}

func (c *mockConn) read() (*protocol.Request, error) {
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
	req.Header = protocol.NewRequestHeaderFromBytes(&headerAsBytes)
	req.Body = make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = c.Read(req.Body)
		if err != nil {
			return nil, err
		}
	}
	return &req, nil
}

func (c *mockConn) writeResponse(resp *protocol.Response) error {
	headerBytes := resp.Header.ToBytes()
	respBytes := make([]byte, protocol.ResponseHeaderSize+len(resp.Body))
	copy(respBytes, headerBytes[:])
	copy(respBytes[protocol.ResponseHeaderSize:], resp.Body)
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

func requestToBytes(req *protocol.Request) []byte {
	headerBytes := req.Header.ToBytes()
	b := make([]byte, protocol.RequestHeaderSize+len(req.Body))
	copy(b, headerBytes[:])
	copy(b[protocol.RequestHeaderSize:], req.Body)
	return b
}

func TestConnectionReadRequest(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &mockConn{Conn: serverConn, cfg: config.Default()}

	key := [protocol.RequestKeySize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	body := []byte("test_body_value")
	req := protocol.NewRequest(action.Set, key, body)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotReq, err := c.read()
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

	_, err := clientConn.Write(requestToBytes(req))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	wg.Wait()
}

func TestConnectionWriteResponse(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &mockConn{Conn: serverConn, cfg: config.Default()}
	resp := protocol.NewResponse(status.OK, []byte("response_body"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := c.writeResponse(resp)
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

	c := &mockConn{Conn: serverConn, cfg: config.Default()}

	key := [protocol.RequestKeySize]byte{0xFF, 0xEE, 0xDD}
	body := []byte("roundtrip_data")
	req := protocol.NewRequest(action.Get, key, body)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotReq, err := c.read()
		if err != nil {
			t.Errorf("read() error: %v", err)
			return
		}
		if gotReq.Header.Operation != uint8(action.Get) {
			t.Errorf("expected operation %d, got %d", action.Get, gotReq.Header.Operation)
		}
		err = c.writeResponse(protocol.NewResponse(status.NoSuchKey, nil))
		if err != nil {
			t.Errorf("writeResponse() error: %v", err)
		}
	}()

	_, err := clientConn.Write(requestToBytes(req))
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

	c := &mockConn{Conn: serverConn, cfg: config.Default()}
	defer serverConn.Close()

	_, err := c.read()
	if err == nil {
		t.Error("expected error reading from closed connection, got nil")
	}
}

func asbytesResponseReader(conn net.Conn) (status.Status, []byte, error) {
	buf := make([]byte, protocol.ResponseHeaderSize)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return 0, nil, err
	}
	bodySize := binary.LittleEndian.Uint32(buf[:4])
	sts := status.Status(buf[4])
	var body []byte
	if bodySize > 0 {
		body = make([]byte, bodySize)
		if _, err := io.ReadFull(conn, body); err != nil {
			return 0, nil, err
		}
	}
	return sts, body, nil
}

func TestConnServe_SetAndGet(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()
	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	key := [protocol.RequestKeySize]byte{1, 2, 3, 4}
	val := []byte("stored_value")

	setReq := protocol.NewRequest(action.Set, key, val)
	_, err := clientConn.Write(requestToBytes(setReq))
	if err != nil {
		t.Fatalf("failed to write SET: %v", err)
	}
	sts, body, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read SET response: %v", err)
	}
	if sts != status.OK {
		t.Fatalf("SET expected OK, got %v", sts)
	}
	if len(body) != 0 {
		t.Fatalf("SET expected empty body, got %q", string(body))
	}

	getReq := protocol.NewRequest(action.Get, key, nil)
	_, err = clientConn.Write(requestToBytes(getReq))
	if err != nil {
		t.Fatalf("failed to write GET: %v", err)
	}
	sts, body, err = asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read GET response: %v", err)
	}
	if sts != status.OK {
		t.Fatalf("GET expected OK, got %v", sts)
	}
	if !bytes.Equal(body, val) {
		t.Errorf("GET body mismatch: got %q, want %q", string(body), string(val))
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_Get_NoSuchKey(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()
	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	key := [protocol.RequestKeySize]byte{0xAA}
	req := protocol.NewRequest(action.Get, key, nil)

	_, err := clientConn.Write(requestToBytes(req))
	if err != nil {
		t.Fatalf("failed to write GET: %v", err)
	}

	sts, _, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if sts != status.NoSuchKey {
		t.Errorf("expected NoSuchKey, got %v", sts)
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_Del(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()

	key := [protocol.RequestKeySize]byte{0xBB}
	database.Set(key, db.Val([]byte("delete_me")))

	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	delReq := protocol.NewRequest(action.Del, key, nil)
	_, err := clientConn.Write(requestToBytes(delReq))
	if err != nil {
		t.Fatalf("failed to write DEL: %v", err)
	}
	sts, _, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read DEL response: %v", err)
	}
	if sts != status.OK {
		t.Errorf("DEL expected OK, got %v", sts)
	}

	getReq := protocol.NewRequest(action.Get, key, nil)
	_, err = clientConn.Write(requestToBytes(getReq))
	if err != nil {
		t.Fatalf("failed to write GET after DEL: %v", err)
	}
	sts, _, err = asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read GET response: %v", err)
	}
	if sts != status.NoSuchKey {
		t.Errorf("GET after DEL expected NoSuchKey, got %v", sts)
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_Exist(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()

	key := [protocol.RequestKeySize]byte{0xCC}
	database.Set(key, db.Val([]byte("exists")))

	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	existReq := protocol.NewRequest(action.Exist, key, nil)
	_, err := clientConn.Write(requestToBytes(existReq))
	if err != nil {
		t.Fatalf("failed to write EXIST: %v", err)
	}
	sts, _, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read EXIST response: %v", err)
	}
	if sts != status.OK {
		t.Errorf("EXIST expected OK, got %v", sts)
	}

	delReq := protocol.NewRequest(action.Del, key, nil)
	_, err = clientConn.Write(requestToBytes(delReq))
	if err != nil {
		t.Fatalf("failed to write DEL: %v", err)
	}
	sts, _, err = asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read DEL response: %v", err)
	}
	if sts != status.OK {
		t.Fatalf("DEL expected OK, got %v", sts)
	}

	existReq2 := protocol.NewRequest(action.Exist, key, nil)
	_, err = clientConn.Write(requestToBytes(existReq2))
	if err != nil {
		t.Fatalf("failed to write EXIST after DEL: %v", err)
	}
	sts, _, err = asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read EXIST response: %v", err)
	}
	if sts != status.NoSuchKey {
		t.Errorf("EXIST after DEL expected NoSuchKey, got %v", sts)
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_InvalidOperation(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()
	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	rawReq := make([]byte, protocol.RequestHeaderSize)
	rawReq[0] = 255

	_, err := clientConn.Write(rawReq)
	if err != nil {
		t.Fatalf("failed to write invalid request: %v", err)
	}

	sts, _, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if sts != status.WrongInput {
		t.Errorf("expected WrongInput, got %v", sts)
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_EOF(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()
	conn := connection.New(serverConn, database, cfg)

	done := make(chan struct{})
	go func() {
		conn.Serve()
		close(done)
	}()

	clientConn.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after client closed connection")
	}

	serverConn.Close()
}

func TestConnServe_ReadTimeout(t *testing.T) {
	os.Setenv("GEDIS_RECEIVE_TIMEOUT", "100")
	os.Setenv("GEDIS_SEND_TIMEOUT", "100")
	defer func() {
		os.Unsetenv("GEDIS_RECEIVE_TIMEOUT")
		os.Unsetenv("GEDIS_SEND_TIMEOUT")
	}()

	cfg := config.Load()
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	_, err := clientConn.Write([]byte{byte(action.Set)})
	if err != nil {
		t.Fatalf("failed to write first byte: %v", err)
	}

	sts, _, err := asbytesResponseReader(clientConn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if sts != status.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", sts)
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}

func TestConnServe_MultipleRequests(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	database := db.NewDB()
	cfg := config.Default()
	conn := connection.New(serverConn, database, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn.Serve()
	}()

	key1 := [protocol.RequestKeySize]byte{1}
	key2 := [protocol.RequestKeySize]byte{2}

	reqs := []struct {
		req    *protocol.Request
		body   []byte
		status status.Status
	}{
		{protocol.NewRequest(action.Set, key1, []byte("val1")), nil, status.OK},
		{protocol.NewRequest(action.Set, key2, []byte("val2")), nil, status.OK},
		{protocol.NewRequest(action.Get, key1, nil), []byte("val1"), status.OK},
		{protocol.NewRequest(action.Get, key2, nil), []byte("val2"), status.OK},
	}

	for _, tc := range reqs {
		_, err := clientConn.Write(requestToBytes(tc.req))
		if err != nil {
			t.Fatalf("failed to write request: %v", err)
		}
		sts, body, err := asbytesResponseReader(clientConn)
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}
		if sts != tc.status {
			t.Errorf("expected status %v, got %v", tc.status, sts)
		}
		if !bytes.Equal(body, tc.body) {
			t.Errorf("expected body %q, got %q", string(tc.body), string(body))
		}
	}

	clientConn.Close()
	serverConn.Close()
	wg.Wait()
}