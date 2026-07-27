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

	db.mu.RLock()
	err = gob.NewEncoder(f).Encode(struct {
		KeyVal map[Key]Val
		KeyTTL map[Key]TTL
	}{
		KeyVal: db.keyVal,
		KeyTTL: db.keyTTL,
	})
	db.mu.RUnlock()

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

	db.mu.Lock()
	db.keyVal = data.KeyVal
	db.keyTTL = data.KeyTTL
	db.mu.Unlock()

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
