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

	for k, v := range db.keyVal {
		v.ttl = v.ttl.decreaseBy1()
		if _, ok := v.ttl.(TTLExpired); ok {
			delete(db.keyVal, k)
		} else {
			db.keyVal[k] = v
		}
	}
}
