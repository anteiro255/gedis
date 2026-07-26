package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) Set(key Key, val Val) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.keyTTL, key)
	db.keyVal[key] = val
}

func (db *DB) Get(key Key) (Val, status.Status) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	val, ok := db.keyVal[key]
	if !ok {
		return Val{}, status.NoSuchKey
	}
	if ttl, ok := db.keyTTL[key]; ok && !ttl.isAlive() {
		return Val{}, status.NoSuchKey
	}
	return val, status.OK
}

func (db *DB) Del(key Key) (sts status.Status) {
	sts = status.OK
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.keyVal[key]; !ok {
		sts = status.NoSuchKey
	} else if ttl, ok := db.keyTTL[key]; ok && !ttl.isAlive() {
		sts = status.NoSuchKey
	}

	delete(db.keyVal, key)
	delete(db.keyTTL, key)
	return
}

func (db *DB) Exists(key Key) status.Status {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if _, ok := db.keyVal[key]; !ok {
		return status.NoSuchKey
	} else if ttl, ok := db.keyTTL[key]; ok && !ttl.isAlive() {
		return status.NoSuchKey
	}

	return status.OK
}
