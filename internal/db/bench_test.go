package db_test

import (
	"testing"

	"github.com/anteiro255/gedis/internal/db"
)

func BenchmarkSet(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Set(key, val)
	}
}

func BenchmarkGet(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))
	database.Set(key, val)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Get(key)
	}
}

func BenchmarkGet_MissingKey(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Get(key)
	}
}

func BenchmarkDel(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Set(key, val)
		database.Del(key)
	}
}

func BenchmarkExists(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))
	database.Set(key, val)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Exists(key)
	}
}

func BenchmarkExists_MissingKey(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Exists(key)
	}
}

func BenchmarkSetGetDelExists(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Set(key, val)
		database.Get(key)
		database.Del(key)
		database.Exists(key)
	}
}

func BenchmarkSetTTL(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))
	database.Set(key, val)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.SetTTL(key, 3600)
	}
}

func BenchmarkGetTTL(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))
	database.Set(key, val)
	database.SetTTL(key, 3600)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.GetTTL(key)
	}
}

func BenchmarkDelTTL(b *testing.B) {
	database := db.NewDB()
	key := db.Key([16]byte{1})
	val := db.Val([]byte("benchmark_value"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		database.Set(key, val)
		database.SetTTL(key, 3600)
		database.DelTTL(key)
	}
}

func BenchmarkParallelSet(b *testing.B) {
	database := db.NewDB()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			var k [16]byte
			k[0] = byte(i)
			i++
			database.Set(db.Key(k), db.Val([]byte("parallel_value")))
		}
	})
}

func BenchmarkParallelGet(b *testing.B) {
	database := db.NewDB()
	var k [16]byte
	k[0] = 1
	database.Set(db.Key(k), db.Val([]byte("parallel_value")))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			database.Get(db.Key(k))
		}
	})
}
