package store

type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Clear()
	List() []string
	This() map[string][]byte
}

type StoreError string

const StoreErrorKeyNotFound StoreError = "key not found"

func (e StoreError) Error() string {
	return string(e)
}
