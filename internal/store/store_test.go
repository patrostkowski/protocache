package store_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/patrostkowski/protocache/internal/store"
	"github.com/patrostkowski/protocache/internal/store/mapstore"
	"github.com/patrostkowski/protocache/internal/store/syncmapstore"
	"github.com/stretchr/testify/assert"
)

func stringToBytes(s string) []byte {
	return []byte(s)
}

func runStoreTests(t *testing.T, name string, newStore func() store.Store) {
	t.Run(name+"/SetAndGet", func(t *testing.T) {
		testKey1 := "key1"
		testValue1 := "value1"
		store := newStore()

		err := store.Set(testKey1, stringToBytes(testValue1))
		assert.NoError(t, err)

		value, err := store.Get(testKey1)
		assert.NoError(t, err)
		assert.Equal(t, stringToBytes(testValue1), value)

		testKey2 := "key2"
		testValue2 := "value2"
		err = store.Set(testKey2, stringToBytes(testValue2))
		assert.NoError(t, err)

		value, err = store.Get(testKey2)
		assert.NoError(t, err)
		assert.Equal(t, stringToBytes(testValue2), value)
	})

	t.Run(name+"/Delete", func(t *testing.T) {
		key := "key1"
		store := newStore()
		_ = store.Set(key, []byte("value1"))
		err := store.Delete(key)
		assert.NoError(t, err)
		_, err = store.Get(key)
		assert.Error(t, err)
	})

	t.Run(name+"/Clear", func(t *testing.T) {
		store := newStore()
		_ = store.Set("key1", []byte("value1"))
		store.Clear()
		_, err := store.Get("key1")
		assert.Error(t, err)
	})

	t.Run(name+"/List", func(t *testing.T) {
		key := "key1"
		store := newStore()
		assert.Equal(t, []string(nil), store.List())

		_ = store.Set(key, []byte("value1"))
		assert.Equal(t, []string{key}, store.List())

		store.Clear()
		assert.Equal(t, []string(nil), store.List())
	})

	t.Run(name+"/This", func(t *testing.T) {
		key := "key1"
		value := []byte("value1")
		store := newStore()
		assert.Equal(t, map[string][]byte{}, store.This())

		_ = store.Set(key, value)
		assert.Equal(t, map[string][]byte{key: value}, store.This())

		store.Clear()
		assert.Equal(t, map[string][]byte{}, store.This())
	})

	t.Run(name+"/ConcurrentSetAndGet", func(t *testing.T) {
		store := newStore()
		numGoroutines := 100
		keys := make([]string, numGoroutines)

		for i := range numGoroutines {
			keys[i] = fmt.Sprintf("key%d", i)
		}

		var wg sync.WaitGroup

		for i := range numGoroutines {
			wg.Add(1)
			go func(key string, value []byte) {
				defer wg.Done()
				assert.NoError(t, store.Set(key, value))
			}(keys[i], []byte(fmt.Sprintf("value%d", i)))
		}

		wg.Wait()

		for i := range numGoroutines {
			wg.Add(1)
			go func(key string, expectedValue string) {
				defer wg.Done()
				value, err := store.Get(key)
				assert.NoError(t, err)
				assert.Equal(t, stringToBytes(expectedValue), value)
			}(keys[i], fmt.Sprintf("value%d", i))
		}

		wg.Wait()
	})
}

func TestStoreImplementations(t *testing.T) {
	runStoreTests(t, "MapStore", func() store.Store {
		return mapstore.NewMapStore()
	})
	runStoreTests(t, "SyncMapStore", func() store.Store {
		return syncmapstore.NewSyncMapStore()
	})
}

func RunBenchmarkSet(b *testing.B, name string, newStore func() store.Store) {
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

func RunBenchmarkGet(b *testing.B, name string, newStore func() store.Store) {
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

func RunBenchmarkDelete(b *testing.B, name string, newStore func() store.Store) {
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

func RunBenchmarkConcurrentSetAndGet(b *testing.B, name string, newStore func() store.Store) {
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
