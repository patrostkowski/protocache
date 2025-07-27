package mapstore_test

import (
	"testing"

	"github.com/patrostkowski/protocache/internal/store"
	"github.com/patrostkowski/protocache/internal/store/mapstore"
)

func BenchmarkMapStore(b *testing.B) {
	store.RunBenchmarkSet(b, "MapStore", func() store.Store {
		return mapstore.NewMapStore()
	})
	store.RunBenchmarkGet(b, "MapStore", func() store.Store {
		return mapstore.NewMapStore()
	})
	store.RunBenchmarkDelete(b, "MapStore", func() store.Store {
		return mapstore.NewMapStore()
	})
	store.RunBenchmarkConcurrentSetAndGet(b, "MapStore", func() store.Store {
		return mapstore.NewMapStore()
	})
}
