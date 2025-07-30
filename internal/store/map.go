// Copyright 2025 Patryk Rostkowski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"maps"
	"slices"
	"sync"

	"github.com/patrostkowski/protocache/internal/logger"
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
			logger.Debug("Evicted key from store", "key", evictKey)
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
	logger.Debug("Cleared entire store")
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
