package db

import (
	"context"
	"math/rand/v2"
	"time"
)

// Only for non-raft mode
// Blocking function
// RunTTLManager starts the background TTL expiry worker.
func (db *DB) RunTTLManager(ctx context.Context, checks uint) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			db.TickTTLs(checks, func(s *Shard, k Key) {
				delete(s.keyVal, k)
				delete(s.keyTTL, k)
			})
		}
	}
}

func (db *DB) TickTTLs(checks uint, del func(s *Shard, k Key)) {
	now := unixNow()

	firstShardIndex := rand.IntN(shardCount)
	i := firstShardIndex
	for {

		s := &db.shards[i]
		s.mu.Lock()
		for k, ttl := range s.keyTTL {
			if checks == 0 {
				s.mu.Unlock()
				return
			}

			if !ttl.isAlive(now) {
				del(s, k)
			}

			checks--
		}
		s.mu.Unlock()
		if checks == 0 {
			break
		}
		i++
		i %= shardCount
		if i == firstShardIndex {
			break
		}
	}
}
