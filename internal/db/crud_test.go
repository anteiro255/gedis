package db

import (
	"testing"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestDB_CRUD(t *testing.T) {
	db := NewDB()
	key := Key([16]byte{1})
	val := &Val{data: []byte("hello")}

	// Test Set
	err := db.Set(key, val)
	if err != status.OK {
		t.Fatalf("expected OK status, got %v", err)
	}

	// Test Set existing key
	err = db.Set(key, val)
	if err == status.OK {
		t.Error("expected SuchKeyAlreadyExists status when setting existing key, got OK status")
	}

	// Test Exists
	if !db.Exists(key) {
		t.Error("expected key to exist")
	}

	// Test Get
	got, err := db.Get(key)
	if err != status.OK || string(got.data) != "hello" {
		t.Errorf("expected 'hello', got %s", string(got.data))
	}

	// Test Del
	err = db.Del(key)
	if err != status.OK {
		t.Fatalf("expected successful deletion, got %v", err)
	}

	// Test Get after Del
	_, err = db.Get(key)
	if err == status.OK {
		t.Error("expected NoSuchKey status, got nil")
	}

	// Test Del non-existent
	err = db.Del(key)
	if err == status.OK {
		t.Error("expected NoSuchKey status when deleting non-existent key, got OK")
	}
}
