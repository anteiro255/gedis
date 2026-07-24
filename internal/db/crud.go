package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) Set(key Key, val *Val) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	existing, ok := db.keyVal[key]
	if ok {
		if _, expired := existing.ttl.(TTLExpired); !expired {
			return status.KeyAlreadyExists
		}
	}
	db.keyVal[key] = *val
	return status.OK
}

func (db *DB) Get(key Key) (Val, status.Status) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	val, ok := db.keyVal[key]
	if !ok {
		return Val{}, status.NoSuchKey
	}
	if _, expired := val.ttl.(TTLExpired); expired {
		return Val{}, status.NoSuchKey
	}
	return val, status.OK
}

func (db *DB) Del(key Key) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	val, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}
	delete(db.keyVal, key)
	if _, expired := val.ttl.(TTLExpired); expired {
		return status.NoSuchKey
	}
	return status.OK
}

func (db *DB) Exists(key Key) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	val, ok := db.keyVal[key]
	if !ok {
		return false
	}
	if _, expired := val.ttl.(TTLExpired); expired {
		return false
	}
	return true
}
