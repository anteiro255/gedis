package db_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestDB_TTLManagerEviction(t *testing.T) {
	database := db.NewDB()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.RunTTLManager(ctx)

	key := db.Key([16]byte{9, 9, 9})
	val := db.Val([]byte("expiring_value"))

	database.Set(key, val)
	// Set TTL to 1 second
	if sts := database.SetTTL(key, 1); sts != status.OK {
		t.Fatalf("SetTTL failed: %v", sts)
	}

	// Verify key exists right away
	if _, sts := database.Get(key); sts != status.OK {
		t.Fatalf("expected key to exist initially, got %v", sts)
	}

	// Wait 2.2 seconds for TTL manager tick to trigger and evict
	time.Sleep(2200 * time.Millisecond)

	if _, sts := database.Get(key); sts != status.NoSuchKey {
		t.Errorf("expected key to be evicted by TTL manager, got status %v", sts)
	}
}

func TestDB_ExpiredKeyAccess(t *testing.T) {
	database := db.NewDB()
	key := db.Key([16]byte{5, 5, 5})
	val := db.Val([]byte("data"))

	database.Set(key, val)
	// Set 1s TTL
	database.SetTTL(key, 1)

	// Wait 1.1s so ttl.isAlive() returns false
	time.Sleep(1100 * time.Millisecond)

	// Get on expired key should return NoSuchKey
	if _, sts := database.Get(key); sts != status.NoSuchKey {
		t.Errorf("Get on expired key expected NoSuchKey, got %v", sts)
	}

	// Del on expired key should return NoSuchKey
	if sts := database.Del(key); sts != status.NoSuchKey {
		t.Errorf("Del on expired key expected NoSuchKey, got %v", sts)
	}

	// Re-Set key and expire it again
	database.Set(key, val)
	database.SetTTL(key, 1)
	time.Sleep(1100 * time.Millisecond)

	// Exists on expired key should return NoSuchKey
	if sts := database.Exists(key); sts != status.NoSuchKey {
		t.Errorf("Exists on expired key expected NoSuchKey, got %v", sts)
	}

	// SetTTL on expired key should clean up and return NoSuchKey
	if sts := database.SetTTL(key, 5); sts != status.NoSuchKey {
		t.Errorf("SetTTL on expired key expected NoSuchKey, got %v", sts)
	}

	// Re-Set key and expire it again for DelTTL test
	database.Set(key, val)
	database.SetTTL(key, 1)
	time.Sleep(1100 * time.Millisecond)

	if sts := database.DelTTL(key); sts != status.NoSuchKey {
		t.Errorf("DelTTL on expired key expected NoSuchKey, got %v", sts)
	}
}

func TestDB_ConcurrentAccess(t *testing.T) {
	database := db.NewDB()
	goroutines := 20
	iterations := 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				var k [16]byte
				k[0] = byte(j % 10)
				key := db.Key(k)
				val := db.Val([]byte(fmt.Sprintf("val-%d-%d", gID, j)))

				database.Set(key, val)
				database.Get(key)
				database.Exists(key)
				database.SetTTL(key, 10)
				database.GetTTL(key)
				database.DelTTL(key)
				if j%2 == 0 {
					database.Del(key)
				}
			}
		}(i)
	}

	wg.Wait()
}
