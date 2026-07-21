package db

import (
	"context"
	"sync"
	"time"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type TTL interface {
	// Return the new TTL state after decreasing
	decreaseBy1() TTL
}

type TTLNever struct{}

func (t TTLNever) decreaseBy1() TTL {
	return t // Never stays Never
}

type TTLExpired struct{}

func (t TTLExpired) decreaseBy1() TTL {
	return t // Expired stays Expired
}

type TTLSeconds struct {
	Seconds uint32
}

func (t TTLSeconds) decreaseBy1() TTL {
	if t.Seconds == 0 {
		return TTLExpired{} // Morph state into Expired
	}

	return TTLSeconds{Seconds: t.Seconds - 1} // Return updated seconds
}

type Key [16]byte
type Val struct {
	data []byte
	ttl  TTL
}

func NewVal(data []byte) *Val {
	return &Val{
		data: data,
		ttl:  TTLNever{},
	}
}
func (v *Val) AddTTL(ttl TTL) *Val {
	return &Val{
		data: v.data,
		ttl:  ttl,
	}
}

type DB struct {
	mu     sync.RWMutex
	keyVal map[Key]Val
}

func NewDB() *DB {
	return &DB{
		keyVal: make(map[Key]Val),
	}
}

func (db *DB) Set(key Key, val Val) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.keyVal[key]
	if ok {
		return status.SUCH_KEY_ALREADY_EXISTS
	}
	db.keyVal[key] = val
	return status.OK
}

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
