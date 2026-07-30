package server_test

import (
	"testing"

	"github.com/anteiro255/go-gedis"
)

func BenchmarkSetGetDelExists(b *testing.B) {
	k := []byte("Hello")
	v := []byte(" world!")
	ctx := b.Context()

	b.ReportAllocs()
	c, err := gedis.NewClient(serverAddr)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		c.Set(ctx, k, v)
		c.Get(ctx, k)
		c.Del(ctx, k)
		c.Exist(ctx, k)
	}
}
