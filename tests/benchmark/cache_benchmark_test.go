// Copyright 2025 Patryk Rostkowski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal_test

import (
	"context"
	"strconv"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/server"
	testhelpers "github.com/patrostkowski/protocache/internal/test"
)

func BenchmarkSet(b *testing.B) {
	cfg := config.DefaultConfig()
	server := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})
	}
}

func BenchmarkGet(b *testing.B) {
	cfg := config.DefaultConfig()
	server := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
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
	cfg := config.DefaultConfig()
	server := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	ctx := context.Background()

	key := "key_to_delete"
	for i := 0; i < b.N; i++ {
		server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})
		server.Delete(ctx, &pb.DeleteRequest{Key: key})
	}
}

func BenchmarkGetParallel(b *testing.B) {
	cfg := config.DefaultConfig()
	server := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
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

func BenchmarkList(b *testing.B) {
	cfg := config.DefaultConfig()
	server := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	ctx := context.Background()

	// Preload some keys
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		server.Set(ctx, &pb.SetRequest{Key: key, Value: []byte("value")})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.List(ctx, &pb.ListRequest{})
	}
}
