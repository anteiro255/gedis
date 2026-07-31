package connection

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol"
	protocolaction "github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func helperSendRequest(t *testing.T, conn net.Conn, req *protocol.Request) {
	t.Helper()
	headerBytes := req.Header.ToBytes()
	if _, err := conn.Write(headerBytes[:]); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if len(req.Body) > 0 {
		if _, err := conn.Write(req.Body); err != nil {
			t.Fatalf("failed to write body: %v", err)
		}
	}
}

func helperReadResponse(t *testing.T, conn net.Conn) *protocol.Response {
	t.Helper()
	var headerBuf [protocol.ResponseHeaderSize]byte
	if _, err := io.ReadFull(conn, headerBuf[:]); err != nil {
		t.Fatalf("failed to read response header: %v", err)
	}
	header := protocol.NewResponseHeaderFromBytes(&headerBuf)
	body := make([]byte, header.BodySize)
	if header.BodySize > 0 {
		if _, err := io.ReadFull(conn, body); err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
	}
	return &protocol.Response{
		Header: header,
		Body:   body,
	}
}

func TestConn_Serve_SetAndGet(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	database := db.NewDB()
	c := New(serverConn, database, config.DefaultServerConfig())

	done := make(chan struct{})
	go func() {
		c.Serve()
		close(done)
	}()

	var key [protocol.RequestKeySize]byte
	copy(key[:], "testkey")
	val := []byte("testvalue")

	// 1. SET
	setReq := protocol.NewRequest(protocolaction.Set, &key, val)
	helperSendRequest(t, clientConn, setReq)
	setRes := helperReadResponse(t, clientConn)
	if setRes.Header.Status != status.OK {
		t.Fatalf("expected SET status OK, got %v", setRes.Header.Status)
	}

	// 2. GET
	getReq := protocol.NewRequest(protocolaction.Get, &key, nil)
	helperSendRequest(t, clientConn, getReq)
	getRes := helperReadResponse(t, clientConn)
	if getRes.Header.Status != status.OK {
		t.Fatalf("expected GET status OK, got %v", getRes.Header.Status)
	}
	if string(getRes.Body) != string(val) {
		t.Fatalf("expected GET body %q, got %q", val, getRes.Body)
	}

	// Close client end and wait for server handler to terminate
	_ = clientConn.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not exit after client close")
	}
}

func TestConn_Serve_ExistAndDel(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	database := db.NewDB()
	c := New(serverConn, database, config.DefaultServerConfig())

	done := make(chan struct{})
	go func() {
		c.Serve()
		close(done)
	}()

	var key [protocol.RequestKeySize]byte
	copy(key[:], "existkey")
	val := []byte("existval")

	// 1. SET
	helperSendRequest(t, clientConn, protocol.NewRequest(protocolaction.Set, &key, val))
	res := helperReadResponse(t, clientConn)
	if res.Header.Status != status.OK {
		t.Fatalf("SET failed: %v", res.Header.Status)
	}

	// 2. EXIST
	helperSendRequest(t, clientConn, protocol.NewRequest(protocolaction.Exist, &key, nil))
	res = helperReadResponse(t, clientConn)
	if res.Header.Status != status.OK {
		t.Fatalf("EXIST failed: %v", res.Header.Status)
	}

	// 3. DEL
	helperSendRequest(t, clientConn, protocol.NewRequest(protocolaction.Del, &key, nil))
	res = helperReadResponse(t, clientConn)
	if res.Header.Status != status.OK {
		t.Fatalf("DEL failed: %v", res.Header.Status)
	}

	// 4. GET after DEL -> NoSuchKey
	helperSendRequest(t, clientConn, protocol.NewRequest(protocolaction.Get, &key, nil))
	res = helperReadResponse(t, clientConn)
	if res.Header.Status != status.NoSuchKey {
		t.Fatalf("expected status.NoSuchKey after DEL, got %v", res.Header.Status)
	}

	_ = clientConn.Close()
	<-done
}

func TestConn_Serve_ClientDisconnectEOF(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	database := db.NewDB()
	c := New(serverConn, database, config.DefaultServerConfig())

	done := make(chan struct{})
	go func() {
		c.Serve()
		close(done)
	}()

	// Immediately close client connection
	_ = clientConn.Close()

	select {
	case <-done:
		// Success: handler terminated cleanly on client EOF
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not exit when client disconnected immediately")
	}
}

func TestConn_Serve_InvalidOperation(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	database := db.NewDB()
	c := New(serverConn, database, config.DefaultServerConfig())

	done := make(chan struct{})
	go func() {
		c.Serve()
		close(done)
	}()

	// Send request header with invalid operation type 255
	var key [protocol.RequestKeySize]byte
	rawHeader := protocol.RequestHeader{
		Operation: 255,
		Key:       key,
		BodySize:  0,
	}
	headerBytes := rawHeader.ToBytes()
	if _, err := clientConn.Write(headerBytes[:]); err != nil {
		t.Fatalf("failed to write raw header: %v", err)
	}

	res := helperReadResponse(t, clientConn)
	if res.Header.Status != status.WrongInput {
		t.Fatalf("expected status.WrongInput for invalid operation, got %v", res.Header.Status)
	}

	_ = clientConn.Close()
	<-done
}
