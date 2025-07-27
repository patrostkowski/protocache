package store

import (
	"fmt"
	"sync"
	"testing"
)

func RunBenchmarkSet(b *testing.B, name string, newStore func() Store) {
	b.Run(name+"/Set", func(b *testing.B) {
		s := newStore()
		for i := 0; i < b.N; i++ {
			err := s.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
			if err != nil {
				b.Fatalf("Failed to set key: %v", err)
			}
		}
	})
}

func RunBenchmarkGet(b *testing.B, name string, newStore func() Store) {
	b.Run(name+"/Get", func(b *testing.B) {
		s := newStore()
		for i := 0; i < b.N; i++ {
			_ = s.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := s.Get(fmt.Sprintf("key%d", i))
			if err != nil {
				b.Fatalf("Failed to get key: %v", err)
			}
		}
	})
}

func RunBenchmarkDelete(b *testing.B, name string, newStore func() Store) {
	b.Run(name+"/Delete", func(b *testing.B) {
		s := newStore()
		for i := 0; i < b.N; i++ {
			_ = s.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := s.Delete(fmt.Sprintf("key%d", i))
			if err != nil {
				b.Fatalf("Failed to delete key: %v", err)
			}
		}
	})
}

func RunBenchmarkConcurrentSetAndGet(b *testing.B, name string, newStore func() Store) {
	b.Run(name+"/ConcurrentSetAndGet", func(b *testing.B) {
		s := newStore()
		numGoroutines := b.N
		keys := make([]string, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			keys[i] = fmt.Sprintf("key%d", i)
		}
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(key string, value []byte) {
				defer wg.Done()
				_ = s.Set(key, value)
			}(keys[i], []byte(fmt.Sprintf("value%d", i)))
		}
		wg.Wait()
		b.ResetTimer()
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				_, _ = s.Get(key)
			}(keys[i])
		}
		wg.Wait()
	})
}
