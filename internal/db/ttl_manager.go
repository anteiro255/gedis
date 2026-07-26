package db

import (
	"context"
	"time"

	"github.com/anteiro255/gedis/internal/config"
)

func (db *DB) RunTTLManager(ctx context.Context, cfg *config.Config) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				db.tickTTLs(cfg.TTLEntriesCheckingPerSecond())
			}
		}
	}()
}

func (db *DB) tickTTLs(i uint) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for k, ttl := range db.keyTTL {
		if i == 0 {
			break
		}

		if !ttl.isAlive() {
			delete(db.keyVal, k)
			delete(db.keyTTL, k)
		}

		i--
	}
}
