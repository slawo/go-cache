package memory

import (
	"fmt"
)

// NewLRU instantiates a LRU cache compatible with the Cache interface.
// It returns an error if the parameeter for capacity is not vaild
func NewLRU[K comparable, D any](capacity int) (*LRU[K, D], error) {
	if capacity < 1 {
		return nil, fmt.Errorf("cannot initialize cache with capacity %d", capacity)
	}
	nodes := make([]llkv[K, D], capacity)
	for i := 1; i < capacity; i++ {
		nodes[i-1].next = &nodes[i]
	}
	return &LRU[K, D]{
		capacity: capacity,
		len:      0,
		index:    make(map[K]*llkv[K, D], capacity),
		free:     &nodes[0],
	}, nil
}

// LRU implements the `Cache` interface and provides an LRU (Least Recently Used)
// cache implementation. It uses a doubly-linked list to store the keys in order of
// their last access time. The `NewLRU` function creates a new LRU cache with the
// specified capacity, and the `Get` and `Put` methods retrieve and insert data into
// the cache, respectively.
type LRU[K comparable, D any] struct {
	capacity int
	len      int
	index    map[K]*llkv[K, D]
	head     *llkv[K, D]
	tail     *llkv[K, D]
	free     *llkv[K, D]
}

// Has reports whether the cache has a key
func (c *LRU[K, D]) Has(key K) (bool, error) {
	_, found := c.index[key]
	return found, nil
}

// Get returns either value for the given key or returns ErrNotFound if
// there is no entry for the given key. The key/value pair is updated
// to the top of the list
func (c *LRU[K, D]) Get(key K) (D, error) {
	var data D
	e, found := c.index[key]
	if !found {
		return data, ErrNotFound
	}
	data = e.data
	c.moveNodeToTop(e)
	return data, nil
}

// Put inserts or updates the given key value pair for the given key
// and puts the pair at the top of the priority list.
//
// When the cache is full the oldest entry is evicted prior to insert.
func (c *LRU[K, D]) Put(key K, data D) error {
	e, found := c.index[key]
	if found {
		e.data = data
	} else {
		if c.capacity <= c.len {
			// repurpose bottom
			e = c.tail
			delete(c.index, e.key)
			e.key = key
			e.data = data
		}
		if e == nil {
			e = c.free
			c.free = e.next
			e.next = nil
			e.key = key
			e.data = data
			c.len++
		}
		c.index[key] = e
	}
	c.moveNodeToTop(e)
	return nil
}

func (c *LRU[K, D]) moveNodeToTop(e *llkv[K, D]) error {
	if e == c.head {
		return nil //is already at the top
	}
	// remove from previous nodes
	if e.next != nil {
		e.next.prev = e.prev
	}
	if e.prev != nil {
		e.prev.next = e.next
	}

	//if was tail
	if c.tail == e {
		c.tail = e.prev
	}
	//if there is a head
	if c.head != nil {
		c.head.prev = e
	}
	e.prev = nil
	e.next = c.head
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
	return nil
}

type llkv[K comparable, D any] struct {
	prev *llkv[K, D]
	next *llkv[K, D]
	key  K
	data D
}
