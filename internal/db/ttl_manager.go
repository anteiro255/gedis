package db

import (
	"context"
	"time"

	"github.com/anteiro255/gedis/internal/config"
)

// Call with "go" for not blocking:
// go db.RunSnapshotter(ctx, cfg)
func (db *DB) RunTTLManager(ctx context.Context, cfg *config.Config) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				db.tickTTLs(cfg.TTLEntryCheckPerSecond())
			}
		}
	}()
}

func (db *DB) tickTTLs(i uint) {
	now := unixNow()
	for shardIndex := range db.shards {
		s := &db.shards[shardIndex]
		s.mu.Lock()
		for k, ttl := range s.keyTTL {
			if i == 0 {
				break
			}

			if !ttl.isAlive(now) {
				delete(s.keyVal, k)
				delete(s.keyTTL, k)
			}

			i--
		}
		s.mu.Unlock()
		if i == 0 {
			break
		}
	}
}
