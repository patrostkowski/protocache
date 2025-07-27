package store

import (
	"maps"
	"slices"
	"sync"
)

type MapStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMapStore() *MapStore {
	return &MapStore{
		data: make(map[string][]byte),
	}
}

func (m *MapStore) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MapStore) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	if !exists {
		return nil, StoreErrorKeyNotFound
	}
	return value, nil
}

func (m *MapStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; !exists {
		return StoreErrorKeyNotFound
	}
	delete(m.data, key)
	return nil
}

func (m *MapStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
}

func (m *MapStore) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Collect(maps.Keys(m.data))
}

func (m *MapStore) This() map[string][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data
}
