package server_test

import (
	"testing"

	"github.com/anteiro255/go-gedis"
)

func BenchmarkSetGetDelExists(b *testing.B) {
	k := []byte("Hello")
	v := []byte(" world!")

	b.ReportAllocs()
	c, err := gedis.NewClient(serverAddr)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		c.Set(k, v)
		c.Get(k)
		c.Del(k)
		c.Exist(k)
	}
}
