package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) Set(key Key, val Val) {
	s := db.shardFor(key)
	s.mu.Lock()

	delete(s.keyTTL, key)
	s.keyVal[key] = val
	s.mu.Unlock()
}

func (db *DB) Get(key Key) (Val, status.Status) {
	s := db.shardFor(key)
	s.mu.RLock()

	val, ok := s.keyVal[key]
	if !ok {
		s.mu.RUnlock()
		return Val{}, status.NoSuchKey
	}
	if ttl, ok := s.keyTTL[key]; ok && !ttl.isAlive(unixNow()) {
		s.mu.RUnlock()
		return Val{}, status.NoSuchKey
	}
	s.mu.RUnlock()
	return val, status.OK
}

func (db *DB) Del(key Key) (sts status.Status) {
	sts = status.OK
	s := db.shardFor(key)
	s.mu.Lock()

	if _, ok := s.keyVal[key]; !ok {
		sts = status.NoSuchKey
	} else if ttl, ok := s.keyTTL[key]; ok && !ttl.isAlive(unixNow()) {
		sts = status.NoSuchKey
	}

	delete(s.keyVal, key)
	delete(s.keyTTL, key)
	s.mu.Unlock()
	return
}

func (db *DB) Exists(key Key) status.Status {
	s := db.shardFor(key)
	s.mu.RLock()

	if _, ok := s.keyVal[key]; !ok {
		s.mu.RUnlock()
		return status.NoSuchKey
	} else if ttl, ok := s.keyTTL[key]; ok && !ttl.isAlive(unixNow()) {
		s.mu.RUnlock()
		return status.NoSuchKey
	}

	s.mu.RUnlock()
	return status.OK
}
