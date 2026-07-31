package db_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestSnapshot(t *testing.T) {
	// There's no snapshot file yet

	key := [16]byte{0, 12, 89, 255}
	val := []byte("ABC123")
	database := db.NewDB()
	database.Set(key, val)
	var snapshot bytes.Buffer
	if err := database.WriteSnapshot(&snapshot); err != nil {
		t.Fatal(err)
	}

	// the snapshot file consists of {key: val}

	database = db.NewDB()
	if err := database.ReadSnapshot(&snapshot); err != nil {
		t.Fatal(err)
	}
	got, stat := database.Get(key)
	if stat != status.OK {
		t.Error("db.Get()", "status", stat.Error(), "expected_status", status.OK.Error())
	}
	if !slices.Equal(got, val) {
		t.Error("db.Get()", "got", got, "expected_got", val)
	}

	database.Del(key)
	snapshot.Reset()
	if err := database.WriteSnapshot(&snapshot); err != nil {
		t.Fatal(err)
	}

	// They snapshot is empty now

	database = db.NewDB()
	if err := database.ReadSnapshot(&snapshot); err != nil {
		t.Fatal(err)
	}
	got, stat = database.Get(key)
	if stat != status.NoSuchKey {
		t.Error("db.Get()", "status", stat.Error(), "expected_status", status.NoSuchKey.Error(), "got", got)
	}

}
