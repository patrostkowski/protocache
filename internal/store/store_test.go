package store

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func storesUnderTest() map[string]Store {
	return map[string]Store{
		"MapStore":     NewMapStore(),
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
