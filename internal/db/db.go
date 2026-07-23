package db

import (
	"sync"
)

// TTL interface
// and structtures that implement it
type TTL interface {
	decreaseBy1() TTL // Return the new TTL state after decreasing
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

// Key and Val structures
// and their methods
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
func (v *Val) SetTTL(ttl TTL) {
	v.ttl = ttl
}
func (v *Val) TTL() TTL {
	return v.ttl
}
func (v *Val) Data() []byte {
	return v.data
}

// The DB structure itself
type DB struct {
	mu     sync.RWMutex
	keyVal map[Key]Val
}

func NewDB() *DB {
	return &DB{
		keyVal: make(map[Key]Val),
	}
}
