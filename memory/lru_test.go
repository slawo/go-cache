package memory_test

import (
	"testing"

	gocache "github.com/slawo/go-cache/memory"
	"github.com/stretchr/testify/assert"
)

func TestLRUInvalidSize(t *testing.T) {
	lRUCache, err := gocache.NewLRU[int, int](0)
	assert.Nil(t, lRUCache)
	assert.EqualError(t, err, "cannot initialize cache with capacity 0")

	lRUCache, err = gocache.NewLRU[int, int](-1)
	assert.Nil(t, lRUCache)
	assert.EqualError(t, err, "cannot initialize cache with capacity -1")

	lRUCache, err = gocache.NewLRU[int, int](1)
	assert.NotNil(t, lRUCache)
	assert.NoError(t, err)
}

func TestLRUImplementsCache(t *testing.T) {
	var cache gocache.Cache[int, int]
	var err error
	cache, err = gocache.NewLRU[int, int](1)
	assert.NotNil(t, cache)
	assert.NoError(t, err)
}

func TestLRU(t *testing.T) {
	var v int
	var err error

	lRUCache, err := gocache.NewLRU[int, int](2)
	assert.NoError(t, err)

	_, err = lRUCache.Get(1) // return 1
	assert.ErrorIs(t, err, gocache.ErrNotFound)

	lRUCache.Put(1, 1)       // cache is {1=1}
	lRUCache.Put(2, 2)       // cache is {1=1, 2=2}
	v, err = lRUCache.Get(1) // return 1
	assert.NoError(t, err)
	assert.Equal(t, 1, v)
	lRUCache.Put(3, 3)       // LRU key was 2, evicts key 2, cache is {1=1, 3=3}
	_, err = lRUCache.Get(2) // returns -1 (not found)
	assert.ErrorIs(t, err, gocache.ErrNotFound)
	lRUCache.Put(4, 4)       // LRU key was 1, evicts key 1, cache is {4=4, 3=3}
	_, err = lRUCache.Get(1) // return -1 (not found)
	assert.ErrorIs(t, err, gocache.ErrNotFound)
	v, err = lRUCache.Get(3) // return 3
	assert.NoError(t, err)
	assert.Equal(t, 3, v)
	v, err = lRUCache.Get(4) // return 4
	assert.NoError(t, err)
	assert.Equal(t, 4, v)
	lRUCache.Put(5, 50)
	lRUCache.Put(6, 60)

	v, err = lRUCache.Get(5) // return 50
	assert.NoError(t, err)
	assert.Equal(t, 50, v)

	v, err = lRUCache.Get(6) // return 60
	assert.NoError(t, err)
	assert.Equal(t, 60, v)

	lRUCache.Put(6, 65)
	v, err = lRUCache.Get(6) // return 65
	assert.NoError(t, err)
	assert.Equal(t, 65, v)
}

func TestLRUSequence(t *testing.T) {
	var v int
	var err error

	lRUCache, err := gocache.NewLRU[int, int](3)
	assert.NoError(t, err)

	lRUCache.Put(1, 1)
	lRUCache.Put(2, 2)
	lRUCache.Put(3, 3)
	lRUCache.Put(4, 4)

	v, err = lRUCache.Get(4)
	assert.NoError(t, err)
	assert.Equal(t, 4, v)

	v, err = lRUCache.Get(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, v)

	v, err = lRUCache.Get(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, v)

	_, err = lRUCache.Get(1)
	assert.EqualError(t, err, "not found")

	lRUCache.Put(5, 5)

	_, err = lRUCache.Get(1)
	assert.EqualError(t, err, "not found")

	v, err = lRUCache.Get(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, v)

	v, err = lRUCache.Get(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, v)

	_, err = lRUCache.Get(4)
	assert.EqualError(t, err, "not found")

	v, err = lRUCache.Get(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, v)
}

func TestLRUGetEmpty(t *testing.T) {
	var err error

	lRUCache, err := gocache.NewLRU[int, int](3)
	assert.NoError(t, err)

	_, err = lRUCache.Get(5)
	assert.EqualError(t, err, "not found")
}
