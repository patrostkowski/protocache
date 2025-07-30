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
	"sync"
)

type SyncMapStore struct {
	data sync.Map
}

func NewSyncMapStore() *SyncMapStore {
	return &SyncMapStore{}
}

func (s *SyncMapStore) Set(key string, value []byte) error {
	s.data.Store(key, value)
	return nil
}

func (s *SyncMapStore) Get(key string) ([]byte, error) {
	val, ok := s.data.Load(key)
	if !ok {
		return nil, StoreErrorKeyNotFound
	}
	return val.([]byte), nil
}

func (s *SyncMapStore) Delete(key string) error {
	_, ok := s.data.Load(key)
	if !ok {
		return StoreErrorKeyNotFound
	}
	s.data.Delete(key)
	return nil
}

func (s *SyncMapStore) Clear() {
	s.data.Range(func(k, _ any) bool {
		s.data.Delete(k)
		return true
	})
}

func (s *SyncMapStore) List() []string {
	var keys []string
	s.data.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func (s *SyncMapStore) This() map[string][]byte {
	snapshot := make(map[string][]byte)
	s.data.Range(func(k, v any) bool {
		snapshot[k.(string)] = v.([]byte)
		return true
	})
	return snapshot
}
