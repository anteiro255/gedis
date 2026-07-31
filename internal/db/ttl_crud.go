package db

import "github.com/anteiro255/gedis/pkg/protocol/status"

func (db *DB) Expire(key Key, deadline TTL) status.Status {
	s := db.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keyVal[key]; !ok {
		return status.NoSuchKey
	}
	if ttl, ok := s.keyTTL[key]; !ok || ttl != deadline {
		return status.NoSuchKey
	}
	delete(s.keyVal, key)
	delete(s.keyTTL, key)
	return status.OK
}

func (db *DB) SetTTL(key Key, ttl TTL) status.Status {
	s := db.shardFor(key)
	s.mu.Lock()

	_, ok := s.keyVal[key]
	if !ok {
		s.mu.Unlock()
		return status.NoSuchKey
	}

	if existingTTL, ok := s.keyTTL[key]; ok && !existingTTL.isAlive(unixNow()) {
		delete(s.keyVal, key)
		delete(s.keyTTL, key)
		s.mu.Unlock()
		return status.NoSuchKey
	}

	s.keyTTL[key] = TTL(unixNow()) + ttl
	s.mu.Unlock()
	return status.OK
}

func (db *DB) GetTTL(key Key) (TTL, status.Status) {
	s := db.shardFor(key)
	s.mu.RLock()

	if _, ok := s.keyVal[key]; !ok {
		s.mu.RUnlock()
		return TTL(0), status.NoSuchKey
	}

	if ttl, ok := s.keyTTL[key]; ok {
		now := unixNow()
		if !ttl.isAlive(now) {
			s.mu.RUnlock()
			return TTL(0), status.NoSuchKey
		}
		rem := ttl - TTL(now)
		s.mu.RUnlock()
		return rem, status.OK
	}
	s.mu.RUnlock()
	return TTL(0), status.NoTTL
}

func (db *DB) DelTTL(key Key) status.Status {
	s := db.shardFor(key)
	s.mu.Lock()

	_, ok := s.keyVal[key]
	if !ok {
		s.mu.Unlock()
		return status.NoSuchKey
	}
	if ttl, ok := s.keyTTL[key]; ok && !ttl.isAlive(unixNow()) {
		delete(s.keyVal, key)
		delete(s.keyTTL, key)
		s.mu.Unlock()
		return status.NoSuchKey
	}
	delete(s.keyTTL, key)
	s.mu.Unlock()
	return status.OK
}
