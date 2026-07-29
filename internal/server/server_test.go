package server_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/server"
	"github.com/anteiro255/gedis/pkg/protocol/status"
	client "github.com/anteiro255/go-gedis"
)

const serverAddr = "127.0.0.1:8080"

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())

	database := db.NewDB()
	database.RunTTLManager(ctx, config.Default())

	s := server.NewServer()
	s.SetDB(database)

	go s.RunAt(ctx, serverAddr)

	time.Sleep(500 * time.Millisecond)

	code := m.Run()
	cancel()
	os.Exit(code)
}

func TestIntegration_SetAndGetMultiple(t *testing.T) {
	c, err := client.NewClient(serverAddr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer c.Close()

	key1 := []byte{1}
	val1 := []byte("val1")

	key2 := []byte{2}
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
	c, err := client.NewClient(serverAddr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer c.Close()

	key := []byte{99}
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

			key := []byte{byte(id)}
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

	buf := make([]byte, 100)
	n, _ := conn.Read(buf)
	if status.Status(n) != status.WrongInput {
		// Server might respond with error response before closing connection
		t.Logf("server sent response to garbage data: %v", buf[:n])
	}
}
