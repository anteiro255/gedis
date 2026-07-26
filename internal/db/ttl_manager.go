package db

import (
	"context"
	"time"
)

func (db *DB) RunTTLManager(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				db.tickTTLs()
			}
		}
	}()
}

func (db *DB) tickTTLs() {
	db.mu.Lock()
	defer db.mu.Unlock()
	for k, ttl := range db.keyTTL {
		if !ttl.isAlive() {
			delete(db.keyVal, k)
			delete(db.keyTTL, k)
		}
	}
}
