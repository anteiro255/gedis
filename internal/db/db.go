package db

import (
	"sync"
	"time"
)

// TTL
type TTL uint32

func (ttl TTL) isAlive() bool {
	return int64(ttl)-int64(time.Now().Unix()) > 0
}

// Key and Val structures
type Key [16]byte
type Val []byte

// The DB structure itself
type DB struct {
	mu     sync.RWMutex
	keyVal map[Key]Val
	keyTTL map[Key]TTL
}

func NewDB() *DB {
	return &DB{
		keyVal: make(map[Key]Val),
		keyTTL: make(map[Key]TTL),
	}
}
