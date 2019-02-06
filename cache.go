package heapcache

import (
	"container/heap"
	"sync"
)

type (
	// Item is something that able to be added to cache.
	Item interface {
		// CacheKey return key of item in cache. It may be any key type (see https://golang.org/ref/spec#KeyType)
		CacheKey() interface{}
		// CacheLess determines priority if items in cache. Items with less priority will be evicted first.
		CacheLess(interface{}) bool
	}

	itemsMap map[interface{}]*wrapper

	// wrapper is a cache item wrapper
	wrapper struct {
		index int
		key   interface{}
		item  Item
	}

	// Cache is a cache abstraction
	Cache struct {
		capacity int
		heap     itemsHeap
		items    itemsMap
		mutex    sync.RWMutex
	}
)

// New creates a new Cache instance
// Capacity allowed to be zero. In this case cache becomes dummy, 'Add' do nothing and items can't be stored in.
func New(capacity int) *Cache {
	if capacity < 0 {
		capacity = 0
	}

	return &Cache{
		capacity: capacity,
		heap:     make(itemsHeap, 0, capacity),
		items:    make(itemsMap, capacity),
	}
}

// Capacity returns capacity of cache
func (c *Cache) Capacity() int {
	return c.capacity
}

// Add adds a `value` into a cache. If `key` already exists, `value` and `priority` will be overwritten.
// `key` must be a KeyType (see https://golang.org/ref/spec#KeyType)
func (c *Cache) Add(items ...Item) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, item := range items {
		c.addItem(item)
	}
}

func (c *Cache) addItem(newItem Item) {
	if c.capacity == 0 {
		return
	}

	key := newItem.CacheKey()

	if item, ok := c.items[key]; ok { // already exists
		c.items[key].item = newItem
		heap.Fix(&c.heap, item.index)
		return
	}

	if len(c.items) >= c.capacity {
		c.evict(1)
	}

	w := wrapper{key: key, item: newItem}

	heap.Push(&c.heap, &w)
	c.items[w.key] = &w
}

// Get gets a value by `key`
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, ok := c.items[key]; ok {
		return item.item, true
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
			heap.Remove(&c.heap, item.index)
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

	c.heap = make(itemsHeap, 0, c.capacity)
	c.items = make(itemsMap, c.capacity)
}

// Evict removes `count` elements with lowest priority.
// TODO Is this useful ever?
func (c *Cache) Evict(count int) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.evict(count)
}

// caller must keep write lock
func (c *Cache) evict(count int) (evicted int) {
	for count > 0 && c.heap.Len() > 0 {
		item := heap.Pop(&c.heap)
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
// Capacity will never be less than zero.
func (c *Cache) SetCapacity(capacity int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.setCapacity(capacity)
}
