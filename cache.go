package heapcache

import (
	"container/heap"
	"sync"
)

// KeyType is a type of item key
type KeyType interface{}

// ValueType is a type of item value
type ValueType interface{}

// PriorityType is a type of item priority
type PriorityType int

// Item is a cache item wrapper
type Item struct {
	index    int
	Key      KeyType
	Value    ValueType
	Priority PriorityType
}

type itemsMap map[KeyType]*Item

// Cache is a cache abstraction
type Cache struct {
	capacity uint
	heap     itemsHeap
	items    itemsMap
	mutex    sync.RWMutex
}

// New creates a new Cache instance
// Capacity allowed to be zero. In this case cache becomes dummy, 'Add' do nothing and items can't be stored in.
func New(capacity uint) *Cache {
	return &Cache{
		capacity: capacity,
		heap:     make(itemsHeap, 0, capacity),
		items:    make(itemsMap, capacity),
	}
}

// Add adds a `value` into a cache. If `key` already exists, `value` and `priority` will be overwritten.
// `key` must be a KeyType (see https://golang.org/ref/spec#KeyType)
func (c *Cache) Add(key KeyType, value ValueType, priority PriorityType) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item := Item{Key: key, Value: value, Priority: priority}
	c.addItem(&item)
}

// AddMany adds many items at once.
func (c *Cache) AddMany(items ...Item) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, item := range items {
		item := item
		c.addItem(&item)
	}
}

func (c *Cache) addItem(newItem *Item) {
	if c.capacity == 0 {
		return
	}

	if item, ok := c.items[newItem.Key]; ok { // already exists
		item.Value = newItem.Value
		if item.Priority != newItem.Priority {
			item.Priority = newItem.Priority
			heap.Fix(&c.heap, item.index)
		}
		return
	}

	if uint(len(c.items)) >= c.capacity {
		c.evict(1)
	}

	heap.Push(&c.heap, newItem)
	c.items[newItem.Key] = newItem
}

// Get gets a value by `key`
func (c *Cache) Get(key KeyType) (ValueType, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, ok := c.items[key]; ok {
		return item.Value, true
	}

	return nil, false
}

// Contains checks of `key` existence.
func (c *Cache) Contains(key KeyType) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, ok := c.items[key]
	return ok
}

// Remove removes a value from cache.
// Returns true if item was removed. Returns false if there is no item in cache.
func (c *Cache) Remove(key KeyType) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items[key]; ok {
		delete(c.items, key)
		heap.Remove(&c.heap, item.index)
		return true
	}

	return false
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

// Evict removes `count` elements with lowest priority
func (c *Cache) Evict(count uint) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return int(c.evict(count))
}

// caller must keep write lock
func (c *Cache) evict(count uint) (evicted uint) {
	for count > 0 && c.heap.Len() > 0 {
		item := heap.Pop(&c.heap)
		delete(c.items, item.(*Item).Key)
		count--
		evicted++
	}
	return
}
