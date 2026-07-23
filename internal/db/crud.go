package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) Set(key Key, val *Val) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.keyVal[key]
	if ok {
		return status.SuchKeyAlreadyExists
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
	return val, status.OK
}
func (db *DB) Del(key Key) status.Status {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.keyVal[key]
	if !ok {
		return status.NoSuchKey
	}
	delete(db.keyVal, key)
	return status.OK
}

func (db *DB) Exists(key Key) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, ok := db.keyVal[key]
	return ok
}
