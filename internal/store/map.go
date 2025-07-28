package store

import (
	"maps"
	"slices"
	"sync"
)

type MapStore struct {
	data             map[string][]byte
	mu               sync.RWMutex
	evictionStrategy EvictionStrategy
}

func NewMapStore(strategy EvictionStrategy) *MapStore {
	return &MapStore{
		data:             make(map[string][]byte),
		evictionStrategy: strategy,
	}
}

func (m *MapStore) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.evictionStrategy != nil {
		if evictKey, shouldEvict := m.evictionStrategy.Evict(m.data); shouldEvict {
			delete(m.data, evictKey)
			m.evictionStrategy.OnDelete(evictKey)
		}
		m.evictionStrategy.OnInsert(key, len(value))
	}

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
	if m.evictionStrategy != nil {
		m.evictionStrategy.OnAccess(key)
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
	if m.evictionStrategy != nil {
		m.evictionStrategy.OnDelete(key)
	}
	return nil
}

func (m *MapStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
	if m.evictionStrategy != nil {
		m.evictionStrategy.Reset()
	}
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
