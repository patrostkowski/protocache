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
	"time"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/logger"
)

const (
	defaultEvictionPolicyCapacity = 1 << 30
)

type EvictionStrategy interface {
	OnAccess(key string)
	OnInsert(key string, size int)
	OnDelete(key string)
	Evict(data map[string][]byte) (evictedKey string, shouldEvict bool)
	Reset()
}

func NewEvictionStrategy(policy v1alpha.EvictionPolicy, capacity int) EvictionStrategy {
	switch policy {
	case v1alpha.EvictionLRU:
		return NewLRUStrategy(capacity)
	case v1alpha.EvictionLFU:
		// return NewLFUStrategy(capacity)
		return nil
	case v1alpha.EvictionRandom:
		// return NewRandomStrategy(capacity)
		return nil
	default:
		return nil
	}
}

type LRUStrategy struct {
	capacity int
	access   map[string]time.Time
}

func NewLRUStrategy(capacity int) *LRUStrategy {
	return &LRUStrategy{
		capacity: capacity,
		access:   make(map[string]time.Time),
	}
}

func (l *LRUStrategy) OnAccess(key string) {
	l.access[key] = time.Now()
}

func (l *LRUStrategy) OnInsert(key string, size int) {
	l.access[key] = time.Now()
}

func (l *LRUStrategy) OnDelete(key string) {
	delete(l.access, key)
}

func (l *LRUStrategy) Evict(data map[string][]byte) (string, bool) {
	if len(data) < l.capacity {
		return "", false
	}

	var oldestKey string
	var oldestTime time.Time

	for k, t := range l.access {
		if _, ok := data[k]; !ok {
			continue // skip stale keys
		}
		if oldestKey == "" || t.Before(oldestTime) {
			oldestKey = k
			oldestTime = t
		}
	}

	logger.Debug("Evicting key (LRU)", "key", oldestKey, "lastAccess", oldestTime)
	return oldestKey, true
}

func (l *LRUStrategy) Reset() {
	l.access = make(map[string]time.Time)
}
