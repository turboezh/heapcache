package cache

import (
	"container/heap"
	"sync"
)

type HeapCacheItem struct {
	index    int
	Key      interface{}
	Value    interface{}
	Priority int64
}

type itemsHeap struct {
	items []*HeapCacheItem
}

func newItemsHeap(capacity uint) *itemsHeap {
	return &itemsHeap{
		items: make([]*HeapCacheItem, 0, capacity),
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
	item := value.(*HeapCacheItem)
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

type HeapCache struct {
	capacity uint
	heap     *itemsHeap
	items    map[interface{}]*HeapCacheItem
	mutex    sync.RWMutex
}

func NewHeapCache(capacity uint) *HeapCache {
	return &HeapCache{
		capacity: capacity,
		heap:     newItemsHeap(capacity),
		items:    make(map[interface{}]*HeapCacheItem, capacity),
	}
}

func (c *HeapCache) Add(key interface{}, value interface{}, priority int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

	item := &HeapCacheItem{Key: key, Value: value, Priority: priority}
	heap.Push(c.heap, item)
	c.items[key] = item
}

func (c *HeapCache) AddMany(items ...HeapCacheItem) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	toAdd := make([]*HeapCacheItem, 0, len(items))

	{
		var oldItem *HeapCacheItem
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

func (c *HeapCache) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if item, ok := c.items[key]; ok {
		return item.Value, true
	}

	return nil, false
}

func (c *HeapCache) Contains(key interface{}) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, ok := c.items[key]
	return ok
}

func (c *HeapCache) Remove(key interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.items[key]; ok {
		delete(c.items, key)
		heap.Remove(c.heap, item.index)
		return true
	}

	return false
}

func (c *HeapCache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// caller must keep write lock
func (c *HeapCache) evict(count uint) {
	for count > 0 {
		item := heap.Pop(c.heap)
		delete(c.items, item.(*HeapCacheItem).Key)
		count--
	}
}
