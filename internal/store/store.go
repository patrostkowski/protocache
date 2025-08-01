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
	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
)

type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Clear()
	List() []string
	This() map[string][]byte
}

func NewStore(engine v1alpha.StoreEngine, policy v1alpha.EvictionPolicy) Store {
	strategy := NewEvictionStrategy(policy, defaultEvictionPolicyCapacity)
	switch engine {
	case v1alpha.SyncMapStoreEngine:
		return NewSyncMapStore()
	default:
		return NewMapStore(strategy)
	}
}
