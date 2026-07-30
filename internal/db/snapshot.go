package db

import (
	"context"
	"encoding/gob"
	"log/slog"
	"os"
	"time"

	"github.com/anteiro255/gedis/internal/config"
)

func (db *DB) SaveSnapshot(path string) error {
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	keyVal := make(map[Key]Val)
	keyTTL := make(map[Key]TTL)
	for i := range db.shards {
		db.shards[i].mu.RLock()
		for key, value := range db.shards[i].keyVal {
			keyVal[key] = value
		}
		for key, ttl := range db.shards[i].keyTTL {
			keyTTL[key] = ttl
		}
	}
	for i := range db.shards {
		db.shards[i].mu.RUnlock()
	}
	err = gob.NewEncoder(f).Encode(struct {
		KeyVal map[Key]Val
		KeyTTL map[Key]TTL
	}{
		KeyVal: keyVal,
		KeyTTL: keyTTL,
	})

	if err != nil {
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
			slog.Debug("No snapshot file was found, creating a new database...", "path", path)
			return nil
		}
		return err
	}
	defer f.Close()

	var data struct {
		KeyVal map[Key]Val
		KeyTTL map[Key]TTL
	}
	if err := gob.NewDecoder(f).Decode(&data); err != nil {
		return err
	}

	for i := range db.shards {
		db.shards[i].mu.Lock()
		db.shards[i].keyVal = make(map[Key]Val)
		db.shards[i].keyTTL = make(map[Key]TTL)
	}
	for key, value := range data.KeyVal {
		s := db.shard(key)
		s.keyVal[key] = value
	}
	for key, ttl := range data.KeyTTL {
		s := db.shard(key)
		s.keyTTL[key] = ttl
	}
	for i := range db.shards {
		db.shards[i].mu.Unlock()
	}

	return nil
}

// Call with "go" for not blocking:
// go db.RunSnapshotter(ctx, cfg)
func (db *DB) RunSnapshotter(ctx context.Context, cfg *config.Config) {
	path := cfg.SnapshotPath()
	interval := cfg.SnapshotInterval()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	save := func() {
		if err := db.SaveSnapshot(path); err != nil {
			slog.Error("Failed to save snapshot", "path", path, "error", err)
		} else {
			slog.Debug("Snapshot saved", "path", path)
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
