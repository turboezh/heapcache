package heapcache

import (
	"container/heap"
	"sync"
)

// Item is a cache item wrapper
type Item struct {
	index    int
	Key      interface{}
	Value    interface{}
	Priority int64
}

type itemsHeap struct {
	items []*Item
}

func newItemsHeap(capacity uint) *itemsHeap {
	return &itemsHeap{
		items: make([]*Item, 0, capacity),
	}
}

func (h *itemsHeap) Len() int {
	return len(h.items)
}

func (h *itemsHeap) Less(i, j int) bool {
	return h.items[i].Priority < h.items[j].Priority
}

func (h *itemsHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

func (h *itemsHeap) Push(value interface{}) {
	item := value.(*Item)
	item.index = len(h.items)
	h.items = append(h.items, item)
}

func (h *itemsHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	h.items = old[0 : n-1]
	return item
}

// Cache is a cache abstraction
type Cache struct {
	capacity uint
	heap     *itemsHeap
	items    map[interface{}]*Item
	mutex    sync.RWMutex
}

// New creates a new Cache instance
// Capacity allowed to be zero. In this case cache becomes dummy, 'Add' do nothing and items can't be stored in.
func New(capacity uint) *Cache {
	return &Cache{
		capacity: capacity,
		heap:     newItemsHeap(capacity),
		items:    make(map[interface{}]*Item, capacity),
	}
}

// Add adds a `value` into a cache. If `key` already exists, `value` and `priority` will be overwritten.
// `key` must be a KeyType (see https://golang.org/ref/spec#KeyType)
func (c *Cache) Add(key interface{}, value interface{}, priority int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.capacity == 0 {
		return
	}

	if item, ok := c.items[key]; ok { // already exists
		item.Value = value
		if item.Priority != priority {
			item.Priority = priority
			heap.Fix(c.heap, item.index)
		}
		return
	}

	if uint(len(c.items)) >= c.capacity {
		c.evict(1)
	}

	item := &Item{Key: key, Value: value, Priority: priority}
	heap.Push(c.heap, item)
	c.items[key] = item
}

// AddMany adds many items at once.
// (It's optimized for this.)
func (c *Cache) AddMany(items ...Item) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.capacity == 0 {
		return
	}

	toAdd := make([]*Item, 0, len(items))

	{
		var oldItem *Item
		var ok bool

		for n := 0; n < len(items); n++ {
			if oldItem, ok = c.items[items[n].Key]; ok { // already exists
				oldItem.Value = items[n].Value
				if oldItem.Priority != items[n].Priority {
					oldItem.Priority = items[n].Priority
					heap.Fix(c.heap, oldItem.index)
				}
			} else {
				toAdd = append(toAdd, &items[n])
			}
		}
	}

	lenItems := uint(len(c.items))
	lenAdd := uint(len(toAdd))
	if lenItems+lenAdd > c.capacity {
		c.evict(lenItems + lenAdd - c.capacity)
	}

	if lenAdd > 0 {
		for _, item := range toAdd {
			c.heap.Push(item)
			c.items[item.Key] = item
		}

		// rebuild heap
		heap.Init(c.heap)
	}
}

// Get gets a value by `key`
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, ok := c.items[key]; ok {
		return item.Value, true
	}

	return nil, false
}

// Contains checks of `key` existence.
func (c *Cache) Contains(key interface{}) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, ok := c.items[key]
	return ok
}

// Remove removes a value from cache.
// Returns true if item was removed. Returns false if there is no item in cache.
func (c *Cache) Remove(key interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items[key]; ok {
		delete(c.items, key)
		heap.Remove(c.heap, item.index)
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

// caller must keep write lock
func (c *Cache) evict(count uint) {
	for count > 0 {
		item := heap.Pop(c.heap)
		delete(c.items, item.(*Item).Key)
		count--
	}
}
