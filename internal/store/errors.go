package store

type StoreError string

const StoreErrorKeyNotFound StoreError = "key not found"

func (e StoreError) Error() string {
	return string(e)
}
