package server_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/server"
	"github.com/anteiro255/gedis/pkg/protocol"
	client "github.com/anteiro255/go-gedis"
)

func startTestServer(t *testing.T) string {
	database := db.NewDB()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	database.RunTTLManager(ctx, config.Default())

	s := server.NewServer()
	s.SetDB(database)

	go s.RunAt(t.Context(), "127.0.0.1:0")

	for range 50 {
		if serverAddr := s.Addr(); serverAddr != "" {
			return serverAddr
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server failed to start listening")
	return ""
}

func TestIntegration_SetAndGetMultiple(t *testing.T) {
	serverAddr := startTestServer(t)

	c, err := client.NewClient(serverAddr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer c.Close()

	key1 := [protocol.RequestKeySize]byte{1}
	val1 := []byte("val1")

	key2 := [protocol.RequestKeySize]byte{2}
	val2 := []byte("val2")

	// 1. Set key1
	err = c.Set(key1, val1)
	if err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}

	// 2. Set key2 (tests multiple requests over single connection)
	err = c.Set(key2, val2)
	if err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}

	// 3. Get key1
	got1, err := c.Get(key1)
	if err != nil {
		t.Fatalf("Get key1 failed: %v", err)
	}
	if string(got1) != string(val1) {
		t.Fatalf("Get key1 expected '%s', got '%s'", val1, got1)
	}

	// 4. Get key2
	got2, err := c.Get(key2)
	if err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}
	if string(got2) != string(val2) {
		t.Fatalf("Get key2 expected '%s', got '%s'", val2, got2)
	}

	// 5. Exist key1
	exists, err := c.Exist(key1)
	if err != nil || !exists {
		t.Fatalf("Exist key1 expected true, got %v, err %v", exists, err)
	}

	// 6. Del key1
	err = c.Del(key1)
	if err != nil {
		t.Fatalf("Del key1 failed: %v", err)
	}

	// 7. Get key1 after Del
	_, err = c.Get(key1)
	if !errors.Is(err, client.ErrNoSuchKey) {
		t.Fatalf("Get key1 after Del expected status.NoSuchKey, got: %v", err)
	}
}

func TestIntegration_TTLOperations(t *testing.T) {
	serverAddr := startTestServer(t)

	c, err := client.NewClient(serverAddr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer c.Close()

	key := [protocol.RequestKeySize]byte{99}
	val := []byte("ttl-val")

	err = c.Set(key, val)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Initially no TTL set
	_, err = c.TTLGet(key)
	if err != client.ErrNoTTL {
		t.Fatalf("expected TTL not to exist before TTLSet, got %v", err.Error())
	}

	// Set TTL to 5 seconds
	err = c.TTLSet(key, 5)
	if err != nil {
		t.Fatalf("TTLSet failed: %v", err)
	}

	secs, err := c.TTLGet(key)
	if err != nil {
		t.Fatalf("TTLGet failed: %v", err)
	}
	if secs != 5 {
		t.Fatalf("expected 5 seconds, got %d", secs)
	}

	// Remove TTL
	err = c.TTLDel(key)
	if err != nil {
		t.Fatalf("TTLDel failed: %v", err)
	}

	_, err = c.TTLGet(key)
	if err != client.ErrNoTTL {
		t.Fatalf("expected TTL not to exist after TTLDel, got %v", err.Error())
	}
}

func TestIntegration_ConcurrentClients(t *testing.T) {
	serverAddr := startTestServer(t)

	numClients := 10
	errChan := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			c, err := client.NewClient(serverAddr)
			if err != nil {
				errChan <- err
				return
			}
			defer c.Close()

			key := [protocol.RequestKeySize]byte{byte(id)}
			val := []byte("concurrent_data")

			if err := c.Set(key, val); err != nil {
				errChan <- err
				return
			}
			got, err := c.Get(key)
			if err != nil || string(got) != string(val) {
				errChan <- err
				return
			}
			errChan <- nil
		}(i)
	}

	for i := 0; i < numClients; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("concurrent client error: %v", err)
		}
	}
}

func TestServer_RawConnection_MalformedData(t *testing.T) {
	serverAddr := startTestServer(t)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	// Send garbage data that does not form a valid request header
	_, err = conn.Write([]byte("garbage_short_data"))
	if err != nil {
		t.Fatalf("failed to write garbage data: %v", err)
	}

	// Server should close connection gracefully on malformed header
	buf := make([]byte, 100)
	n, _ := conn.Read(buf)
	if n > 0 {
		// Server might respond with error response before closing connection
		t.Logf("server sent response to garbage data: %v", buf[:n])
	}
}
