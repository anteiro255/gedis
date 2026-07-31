package db

import (
	"context"
	"encoding/gob"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type snapshotData struct {
	KeyVal map[Key]Val
	KeyTTL map[Key]TTL
}

// collectSnapshot takes a consistent read-locked view of all database shards.
// The locks are released before encoding so serialization does not block
// normal database operations.
func (db *DB) collectSnapshot() snapshotData {
	data := snapshotData{KeyVal: make(map[Key]Val), KeyTTL: make(map[Key]TTL)}
	for i := range db.shards {
		db.shards[i].mu.RLock()
		for key, value := range db.shards[i].keyVal {
			data.KeyVal[key] = value
		}
		for key, ttl := range db.shards[i].keyTTL {
			data.KeyTTL[key] = ttl
		}
	}
	for i := range db.shards {
		db.shards[i].mu.RUnlock()
	}
	return data
}

// WriteSnapshot serializes the database for the Raft FSM snapshot.
func (db *DB) WriteSnapshot(w io.Writer) error {
	return gob.NewEncoder(w).Encode(db.collectSnapshot())
}

// ReadSnapshot restores all keys and TTL metadata from a serialized snapshot.
func (db *DB) ReadSnapshot(r io.Reader) error {
	var data snapshotData
	if err := gob.NewDecoder(r).Decode(&data); err != nil {
		return err
	}
	for i := range db.shards {
		db.shards[i].mu.Lock()
		db.shards[i].keyVal = make(map[Key]Val)
		db.shards[i].keyTTL = make(map[Key]TTL)
	}
	for key, value := range data.KeyVal {
		db.shardFor(key).keyVal[key] = value
	}
	for key, ttl := range data.KeyTTL {
		db.shardFor(key).keyTTL[key] = ttl
	}
	for i := range db.shards {
		db.shards[i].mu.Unlock()
	}
	return nil
}

// SaveSnapshot atomically replaces the standalone-mode database snapshot.
func (db *DB) SaveSnapshot(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if err := db.WriteSnapshot(f); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func (db *DB) LoadSnapshot(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("No standalone snapshot was found", "path", path)
			return nil
		}
		return err
	}
	defer f.Close()
	return db.ReadSnapshot(f)
}

// RunSnapshotter periodically persists the database for standalone mode.
func (db *DB) RunSnapshotter(ctx context.Context, path string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	save := func() {
		if err := db.SaveSnapshot(path); err != nil {
			slog.Error("Failed to save standalone snapshot", "path", path, "error", err)
		}
	}
	for {
		select {
		case <-ctx.Done():
			save()
			return
		case <-ticker.C:
			save()
		}
	}
}
