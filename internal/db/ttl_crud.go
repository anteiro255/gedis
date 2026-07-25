package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) SetTTL(key Key, ttl TTL) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	val, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}
	if _, expired := val.ttl.(TTLExpired); expired {
		return status.NoSuchKey
	}

	val.ttl = ttl
	db.keyVal[key] = val
	return status.OK
}

func (db *DB) GetTTL(key Key) (TTL, status.Status) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	val, ok := db.keyVal[key]
	if !ok {
		return nil, status.NoSuchKey
	}
	if _, expired := val.ttl.(TTLExpired); expired {
		return nil, status.NoSuchKey
	}

	return val.ttl, status.OK
}

func (db *DB) DelTTL(key Key) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	val, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}
	if _, expired := val.ttl.(TTLExpired); expired {
		return status.NoSuchKey
	}

	val.ttl = TTLNever{}
	db.keyVal[key] = val
	return status.OK
}

func (db *DB) ExistsTTL(key Key) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	val, ok := db.keyVal[key]
	if !ok {
		return false
	}
	_, isTTL := val.ttl.(TTLSeconds)
	return isTTL
}
