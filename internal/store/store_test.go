package store

import (
	"sort"
	"testing"
	"time"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/stretchr/testify/assert"
)

func storesUnderTest() map[string]Store {
	return map[string]Store{
		"MapStore":     NewMapStore(nil),
		"SyncMapStore": NewSyncMapStore(),
	}
}

func TestStore_Set(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			err := store.Set("foo", []byte("bar"))
			assert.NoError(t, err)
		})
	}
}

func TestStore_Get(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("foo", []byte("bar"))
			got, err := store.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, []byte("bar"), got)
		})
	}
}

func TestStore_Delete(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("foo", []byte("bar"))
			err := store.Delete("foo")
			assert.NoError(t, err)

			_, err = store.Get("foo")
			assert.Error(t, err)
		})
	}
}

func TestStore_List(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			store.Clear()
			_ = store.Set("a", []byte("1"))
			_ = store.Set("b", []byte("2"))

			keys := store.List()
			sort.Strings(keys)

			assert.Equal(t, []string{"a", "b"}, keys)
		})
	}
}

func TestStore_This(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			store.Clear()
			_ = store.Set("a", []byte("1"))
			_ = store.Set("b", []byte("2"))

			snapshot := store.This()
			assert.Len(t, snapshot, 2)
			assert.Equal(t, []byte("1"), snapshot["a"])
			assert.Equal(t, []byte("2"), snapshot["b"])
		})
	}
}

func TestStore_Clear(t *testing.T) {
	for name, store := range storesUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("x", []byte("123"))
			store.Clear()
			assert.Empty(t, store.List())
		})
	}
}

func lruStoresUnderTest() map[string]Store {
	strategy := NewEvictionStrategy(v1alpha.EvictionLRU, 3)
	return map[string]Store{
		"MapStore": NewMapStore(strategy),
	}
}

func TestMapStore_LRUEviction(t *testing.T) {
	for name, store := range lruStoresUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("a", []byte("1"))
			time.Sleep(1 * time.Millisecond)
			_ = store.Set("b", []byte("2"))
			time.Sleep(1 * time.Millisecond)
			_ = store.Set("c", []byte("3"))

			_, _ = store.Get("a")

			_ = store.Set("d", []byte("4"))

			_, err := store.Get("b")
			assert.Error(t, err, "b should have been evicted (LRU)")

			_, err = store.Get("a")
			assert.NoError(t, err)

			_, err = store.Get("c")
			assert.NoError(t, err)

			_, err = store.Get("d")
			assert.NoError(t, err)
		})
	}
}

func TestMapStore_LRUOnDelete(t *testing.T) {
	for name, store := range lruStoresUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("a", []byte("1"))
			_ = store.Set("b", []byte("2"))
			_ = store.Delete("a")

			_ = store.Set("c", []byte("3"))
			_ = store.Set("d", []byte("4"))

			keys := store.List()
			assert.Len(t, keys, 3)
			assert.ElementsMatch(t, []string{"c", "d", "b"}, keys, "unexpected keys after insert")

			_ = store.Set("e", []byte("5"))
			keys = store.List()
			assert.Len(t, keys, 3)
			assert.Subset(t, keys, []string{"d", "e"}, "d and e should remain, plus one of b/c")
		})
	}
}

func TestMapStore_LRUResetOnClear(t *testing.T) {
	for name, store := range lruStoresUnderTest() {
		t.Run(name, func(t *testing.T) {
			_ = store.Set("x", []byte("1"))
			_ = store.Set("y", []byte("2"))
			store.Clear()

			_ = store.Set("a", []byte("1"))
			_ = store.Set("b", []byte("2"))
			_ = store.Set("c", []byte("3"))

			_ = store.Set("d", []byte("4"))

			keys := store.List()
			assert.Len(t, keys, 3)
			assert.Contains(t, keys, "d")
		})
	}
}
