package store

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stringToBytes(s string) []byte {
	return []byte(s)
}

func TestSetAndGet(t *testing.T) {
	testKey1 := "key1"
	testValue1 := "value1"
	store := NewMapStore()

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
}

func TestDelete(t *testing.T) {
	key := "key1"
	store := NewMapStore()

	err := store.Set(key, []byte("value1"))
	assert.NoError(t, err)

	err = store.Delete(key)
	assert.NoError(t, err)

	_, err = store.Get(key)
	assert.Error(t, err)
}

func TestGetNonExistentKey(t *testing.T) {
	store := NewMapStore()

	_, err := store.Get("nonexistentkey")
	assert.Error(t, err)
}

func TestDeleteNonExistentKey(t *testing.T) {
	store := NewMapStore()

	err := store.Delete("nonexistentkey")
	assert.Error(t, err)
}

func TestConcurrentSetAndGet(t *testing.T) {
	store := NewMapStore()
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
			err := store.Set(key, value)
			assert.NoError(t, err)
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
}

func TestClear(t *testing.T) {
	store := NewMapStore()
	err := store.Set("key1", []byte("value1"))
	assert.NoError(t, err)

	_, err = store.Get("key1")
	assert.NoError(t, err)

	store.Clear()
	_, err = store.Get("key1")
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	key := "key1"
	store := NewMapStore()
	items := store.List()
	assert.Equal(t, []string(nil), items)

	err := store.Set(key, []byte("value1"))
	assert.NoError(t, err)
	items = store.List()
	assert.Equal(t, []string{key}, items)

	store.Clear()
	items = store.List()
	assert.Equal(t, []string(nil), items)
}

func TestThis(t *testing.T) {
	key := "key1"
	value := "value1"
	expectedStore := map[string][]byte{
		key: []byte(value),
	}

	store := NewMapStore()
	items := store.This()
	assert.Equal(t, map[string][]byte{}, items)

	err := store.Set(key, []byte(value))
	assert.NoError(t, err)
	items = store.This()
	assert.Equal(t, expectedStore, items)

	store.Clear()
	items = store.This()
	assert.Equal(t, map[string][]byte{}, items)
}
