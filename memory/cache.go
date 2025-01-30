package memory

import "errors"

const (
	NotFoundError = "not found"
)

var (
	ErrNotFound = errors.New(NotFoundError)
)

// Cache provides
type Cache[K comparable, D any] interface {
	Has(key K) (bool, error)
	Get(key K) (D, error)
	Put(key K, data D) error
}
