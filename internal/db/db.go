package db

import (
	"sync"
	"time"
)

// TTL
type TTL uint32

type ExpiredEntry struct {
	Key      Key
	Deadline TTL
}

func (ttl TTL) isAlive(now uint32) bool {
	return ttl > TTL(now)
}

// Key and Val structures
type Key [16]byte
type Val []byte

const shardCount = 32

type Shard struct {
	mu     sync.RWMutex
	keyVal map[Key]Val
	keyTTL map[Key]TTL
}

// The DB structure itself
type DB struct {
	shards [shardCount]Shard
}

func NewDB() *DB {
	db := &DB{}
	for i := range db.shards {
		db.shards[i].keyVal = make(map[Key]Val)
		db.shards[i].keyTTL = make(map[Key]TTL)
	}
	return db
}

func (db *DB) shardFor(key Key) *Shard {
	return &db.shards[key[0]%shardCount]
}

func unixNow() uint32 {
	return uint32(time.Now().Unix())
}
