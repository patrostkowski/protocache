package internal_test

import (
	"context"
	"strconv"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal"
)

func BenchmarkSet(b *testing.B) {
	server := internal.NewServer()
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})
	}
}

func BenchmarkGet(b *testing.B) {
	server := internal.NewServer()
	ctx := context.Background()

	// Preload a key
	key := "benchmark_key"
	server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.Get(ctx, &pb.GetRequest{Key: key})
	}
}

func BenchmarkDelete(b *testing.B) {
	server := internal.NewServer()
	ctx := context.Background()

	key := "key_to_delete"
	for i := 0; i < b.N; i++ {
		server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})
		server.Delete(ctx, &pb.DeleteRequest{Key: key})
	}
}

func BenchmarkGetParallel(b *testing.B) {
	server := internal.NewServer()
	ctx := context.Background()
	key := "hot"
	server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("hit")})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			server.Get(ctx, &pb.GetRequest{Key: key})
		}
	})
}
