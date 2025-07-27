package store

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkSet(b *testing.B) {
	store := NewMapStore()
	for i := 0; i < b.N; i++ {
		err := store.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
		if err != nil {
			b.Fatalf("Failed to set key: %v", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	store := NewMapStore()
	for i := 0; i < b.N; i++ {
		err := store.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
		if err != nil {
			b.Fatalf("Failed to set key: %v", err)
		}
	}
	b.ResetTimer() // Reset the timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_, err := store.Get(fmt.Sprintf("key%d", i))
		if err != nil {
			b.Fatalf("Failed to get key: %v", err)
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	store := NewMapStore()
	for i := 0; i < b.N; i++ {
		err := store.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("value%d", i)))
		if err != nil {
			b.Fatalf("Failed to set key: %v", err)
		}
	}
	b.ResetTimer() // Reset the timer to exclude setup time
	for i := 0; i < b.N; i++ {
		err := store.Delete(fmt.Sprintf("key%d", i))
		if err != nil {
			b.Fatalf("Failed to delete key: %v", err)
		}
	}
}

func BenchmarkConcurrentSetAndGet(b *testing.B) {
	store := NewMapStore()
	numGoroutines := b.N
	keys := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	var wg sync.WaitGroup

	// Set keys concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(key string, value []byte) {
			defer wg.Done()
			err := store.Set(key, value)
			if err != nil {
				b.Errorf("Failed to set key %s: %v", key, err)
			}
		}(keys[i], []byte(fmt.Sprintf("value%d", i)))
	}
	wg.Wait()
	b.ResetTimer() // Reset the timer to exclude setup time

	// Get keys concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, err := store.Get(key)
			if err != nil {
				b.Errorf("Failed to get key %s: %v", key, err)
			}
		}(keys[i])
	}
	wg.Wait()
}
