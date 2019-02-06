package heapcache

import (
	"container/heap"
	"sync"
)

type (
	// Less is a user-defined function that compares items to evaluate their priorities.
	// Must return true if `a < b`.
	Less func(a, b interface{}) bool

	// Cache is a cache abstraction.
	// It uses user-defined comparator to evaluate priorities of cached items.
	// Items with lowest priorities will be evicted first.
	Cache struct {
		capacity int
		cmp      Less
		heap     heap.Interface
		items    itemsMap
		mutex    sync.RWMutex
	}

	// wrapper is a cache value wrapper
	wrapper struct {
		index int
		key   interface{}
		value interface{}
	}

	itemsMap map[interface{}]*wrapper
)

// New creates a new Cache instance.
// Capacity allowed to be zero. In this case cache becomes dummy, 'Add' do nothing and items can't be stored in.
func New(capacity int, cmp Less) *Cache {
	if capacity < 0 {
		capacity = 0
	}

	return &Cache{
		capacity: capacity,
		cmp:      cmp,
		heap:     newHeap(capacity, cmp),
		items:    make(itemsMap, capacity),
	}
}

// Capacity returns capacity of cache
func (c *Cache) Capacity() int {
	return c.capacity
}

// Add adds a `value` into a cache. If `key` already exists, `value` will be overwritten.
// `key` must be a KeyType (see https://golang.org/ref/spec#KeyType)
func (c *Cache) Add(key interface{}, value interface{}) {
	if c.capacity == 0 {
		return
	}

	if item, ok := c.items[key]; ok { // already exists
		c.items[key].value = value
		heap.Fix(c.heap, item.index)
		return
	}

	if len(c.items) >= c.capacity {
		c.evict(1)
	}

	w := wrapper{key: key, value: value}

	heap.Push(c.heap, &w)
	c.items[w.key] = &w
}

// Get gets a value by `key`
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, ok := c.items[key]; ok {
		return item.value, true
	}
	return nil, false
}

// All checks if ALL `keys` exists
func (c *Cache) All(keys ...interface{}) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, key := range keys {
		if _, ok := c.items[key]; !ok {
			return false
		}
	}
	return true
}

// Any checks if ANY of `keys` exists
func (c *Cache) Any(keys ...interface{}) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, key := range keys {
		if _, ok := c.items[key]; ok {
			return true
		}
	}
	return false
}

// Remove removes values by keys
// Returns number of actually removed items
func (c *Cache) Remove(keys ...interface{}) (removed int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, key := range keys {
		if item, ok := c.items[key]; ok {
			delete(c.items, key)
			heap.Remove(c.heap, item.index)
			removed++
		}
	}
	return
}

// Len returns a number of items in cache
func (c *Cache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// Purge removes all items
func (c *Cache) Purge() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.heap = newHeap(c.capacity, c.cmp)
	c.items = make(itemsMap, c.capacity)
}

// Evict removes `count` elements with lowest priority.
func (c *Cache) Evict(count int) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.evict(count)
}

// caller must keep write lock
func (c *Cache) evict(count int) (evicted int) {
	for count > 0 && c.heap.Len() > 0 {
		item := heap.Pop(c.heap)
		delete(c.items, item.(*wrapper).key)
		count--
		evicted++
	}
	return
}

// ChangeCapacity change cache capacity by `delta`.
// If `delta` is positive cache capacity will be expanded, if `delta` is negative, it will be shrunk.
// Redundant items will be evicted.
func (c *Cache) ChangeCapacity(delta int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.setCapacity(c.capacity + delta)
}

func (c *Cache) setCapacity(capacity int) {
	if capacity == c.capacity {
		return
	}

	if capacity < 0 {
		capacity = 0
	}

	redundant := len(c.items) - capacity
	if redundant > 0 {
		c.evict(redundant)
	}

	c.capacity = capacity
}

// SetCapacity sets cache capacity.
// Redundant items will be evicted.
// Capacity never become less than zero.
func (c *Cache) SetCapacity(capacity int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.setCapacity(capacity)
}

// Fix calls heap.Init to reorder heap.
func (c *Cache) Fix() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	heap.Init(c.heap)
}
