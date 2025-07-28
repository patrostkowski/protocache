package store

import (
	"fmt"
	"testing"
)

func benchmarkStoreSet(b *testing.B, store Store) {
	keyPrefix := "key"
	value := []byte("value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%s%d", keyPrefix, i)
		_ = store.Set(key, value)
	}
}

func benchmarkStoreGet(b *testing.B, store Store) {
	keyPrefix := "key"
	value := []byte("value")

	for i := 0; i < b.N; i++ {
		_ = store.Set(fmt.Sprintf("%s%d", keyPrefix, i), value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(fmt.Sprintf("%s%d", keyPrefix, i))
	}
}

func benchmarkStoreDelete(b *testing.B, store Store) {
	keyPrefix := "key"
	value := []byte("value")

	for i := 0; i < b.N; i++ {
		_ = store.Set(fmt.Sprintf("%s%d", keyPrefix, i), value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Delete(fmt.Sprintf("%s%d", keyPrefix, i))
	}
}

func BenchmarkMapStore_Set(b *testing.B) {
	benchmarkStoreSet(b, NewMapStore(nil))
}

func BenchmarkSyncMapStore_Set(b *testing.B) {
	benchmarkStoreSet(b, NewSyncMapStore())
}

func BenchmarkMapStore_Get(b *testing.B) {
	benchmarkStoreGet(b, NewMapStore(nil))
}

func BenchmarkSyncMapStore_Get(b *testing.B) {
	benchmarkStoreGet(b, NewSyncMapStore())
}

func BenchmarkMapStore_Delete(b *testing.B) {
	benchmarkStoreDelete(b, NewMapStore(nil))
}

func BenchmarkSyncMapStore_Delete(b *testing.B) {
	benchmarkStoreDelete(b, NewSyncMapStore())
}
