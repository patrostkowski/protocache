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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUEviction_EvictsLeastRecentlyUsed(t *testing.T) {
	capacity := 3
	strategy := NewLRUStrategy(capacity)
	data := make(map[string][]byte)

	// Insert 3 keys
	for i, k := range []string{"a", "b", "c"} {
		key := k
		strategy.OnInsert(key, 1)
		data[key] = []byte{byte(i)}
		// Sleep to make timestamps different
		time.Sleep(10 * time.Millisecond)
	}

	// Access "b" to make it more recent
	strategy.OnAccess("b")

	// Insert a 4th key, should trigger eviction
	strategy.OnInsert("d", 1)
	data["d"] = []byte{3}

	evictKey, shouldEvict := strategy.Evict(data)
	assert.True(t, shouldEvict)
	assert.Equal(t, "a", evictKey) // "a" was least recently used

	strategy.OnDelete(evictKey)
	delete(data, evictKey)

	assert.Len(t, data, 3)
	assert.NotContains(t, data, "a")
}

func TestLRUStrategy_Reset(t *testing.T) {
	strategy := NewLRUStrategy(2)
	strategy.OnInsert("x", 1)
	strategy.OnAccess("x")

	strategy.Reset()

	evictKey, shouldEvict := strategy.Evict(map[string][]byte{
		"x": []byte("value"),
	})
	assert.False(t, shouldEvict)
	assert.Equal(t, "", evictKey)
}
