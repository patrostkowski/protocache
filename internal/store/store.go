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
