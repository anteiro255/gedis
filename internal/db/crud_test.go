package db

import (
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestDB_CRUD(t *testing.T) {
	db := NewDB()
	key := Key([16]byte{1})
	val := Val([]byte("hello"))

	db.Set(key, val)
	if db.Exists(key) == status.NoSuchKey {
		t.Error("expected key to exist")
	}

	// Test Get
	got, sts := db.Get(key)
	if sts != status.OK {
		t.Errorf("Expected status=OK, got status=%v", sts)
	}
	if string(got) != "hello" {
		t.Errorf("expected 'hello', got %s", string(got))
	}

	db.Del(key)
	// Test Get after Del
	_, sts = db.Get(key)
	if sts == status.OK {
		t.Error("expected NoSuchKey status, got nil")
	}

	db.Del(key)
	// Test Del non-existent
	db.Del(key)
}

func TestDB_TTL(t *testing.T) {
	db := NewDB()
	key := Key([16]byte{2})
	val := Val([]byte("ttl-test"))

	// Non-existent key SetTTL
	if sts := db.SetTTL(key, 10); sts != status.NoSuchKey {
		t.Errorf("expected NoSuchKey when setting TTL on non-existent key, got %v", sts)
	}

	db.Set(key, val)

	// No TTL set yet
	if _, sts := db.GetTTL(key); sts != status.NoTTL {
		t.Errorf("expected NoTTL status, got %v", sts)
	}

	// Set TTL
	if sts := db.SetTTL(key, 5); sts != status.OK {
		t.Errorf("expected OK on SetTTL, got %v", sts)
	}

	ttl, sts := db.GetTTL(key)
	if sts != status.OK {
		t.Errorf("expected OK on GetTTL, got %v", sts)
	}
	if ttl != 5 {
		t.Errorf("expected TTL remaining ~5, got %d", ttl)
	}

	// Del TTL
	if sts := db.DelTTL(key); sts != status.OK {
		t.Errorf("expected OK on DelTTL, got %v", sts)
	}

	if _, sts := db.GetTTL(key); sts != status.NoTTL {
		t.Errorf("expected NoTTL after DelTTL, got %v", sts)
	}
}
