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

package server

import (
	"context"
	"strconv"
	"testing"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/config"
)

func BenchmarkSet(b *testing.B) {
	cfg := config.DefaultConfig()
	server := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: key, Value: []byte("value")}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	cfg := config.DefaultConfig()
	server := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	ctx := context.Background()

	// Preload a key
	key := "benchmark_key"
	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: key, Value: []byte("value")}); err != nil {
		b.Fatalf("Set failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := server.Get(ctx, &cachev1alpha.GetRequest{Key: key}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	cfg := config.DefaultConfig()
	server := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	ctx := context.Background()

	key := "key_to_delete"
	for i := 0; i < b.N; i++ {
		if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: key, Value: []byte("value")}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
		if _, err := server.Delete(ctx, &cachev1alpha.DeleteRequest{Key: key}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkGetParallel(b *testing.B) {
	cfg := config.DefaultConfig()
	server := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	ctx := context.Background()
	key := "hot"
	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: key, Value: []byte("hit")}); err != nil {
		b.Fatalf("Set failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if _, err := server.Get(ctx, &cachev1alpha.GetRequest{Key: key}); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		}
	})
}

func BenchmarkList(b *testing.B) {
	cfg := config.DefaultConfig()
	server := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	ctx := context.Background()

	// Preload some keys
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: key, Value: []byte("value")}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := server.List(ctx, &cachev1alpha.ListRequest{}); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}
