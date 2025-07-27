package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	store := NewMapStore()

	err := store.Set("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	value, err := store.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value)
	}

	err = store.Set("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	value, err = store.Get("key2")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if string(value) != "value2" {
		t.Errorf("Expected 'value2', got '%s'", value)
	}
}

func TestDelete(t *testing.T) {
	store := NewMapStore()

	err := store.Set("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	err = store.Delete("key1")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = store.Get("key1")
	if err == nil {
		t.Errorf("Expected an error when getting a deleted key, but got none")
	}
}

func TestGetNonExistentKey(t *testing.T) {
	store := NewMapStore()

	_, err := store.Get("nonexistentkey")
	if err == nil {
		t.Errorf("Expected an error when getting a non-existent key, but got none")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	store := NewMapStore()

	err := store.Delete("nonexistentkey")
	if err == nil {
		t.Errorf("Expected an error when deleting a non-existent key, but got none")
	}
}

func TestConcurrentSetAndGet(t *testing.T) {
	store := NewMapStore()
	numGoroutines := 100
	keys := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(key string, value []byte) {
			defer wg.Done()
			err := store.Set(key, value)
			if err != nil {
				t.Errorf("Failed to set key %s: %v", key, err)
			}
		}(keys[i], []byte(fmt.Sprintf("value%d", i)))
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(key string, expectedValue string) {
			defer wg.Done()
			value, err := store.Get(key)
			if err != nil {
				t.Errorf("Failed to get key %s: %v", key, err)
			}
			if string(value) != expectedValue {
				t.Errorf("Expected '%s', got '%s' for key %s", expectedValue, string(value), key)
			}
		}(keys[i], fmt.Sprintf("value%d", i))
	}

	wg.Wait()
}
