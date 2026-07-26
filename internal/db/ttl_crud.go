package db

import (
	"time"

	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func (db *DB) SetTTL(key Key, ttl TTL) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}

	if existingTTL, ok := db.keyTTL[key]; ok && !existingTTL.isAlive() {
		delete(db.keyVal, key)
		delete(db.keyTTL, key)
		return status.NoSuchKey
	}

	db.keyTTL[key] = TTL(time.Now().Unix() + int64(ttl))
	return status.OK
}

func (db *DB) GetTTL(key Key) (TTL, status.Status) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if _, ok := db.keyVal[key]; !ok {
		return TTL(0), status.NoSuchKey
	}

	if ttl, ok := db.keyTTL[key]; ok {
		if !ttl.isAlive() {
			return TTL(0), status.NoSuchKey
		}
		rem := int64(ttl) - time.Now().Unix()
		return TTL(rem), status.OK
	}
	return TTL(0), status.NoTTL
}

func (db *DB) DelTTL(key Key) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}
	if ttl, ok := db.keyTTL[key]; ok && !ttl.isAlive() {
		delete(db.keyVal, key)
		delete(db.keyTTL, key)
		return status.NoSuchKey
	}
	delete(db.keyTTL, key)
	return status.OK
}
