package cache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeapCache_Add(t *testing.T) {
	c := NewHeapCache(10)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 1)

	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, "bar1", val)
	}
	{
		val, ok := c.Get("foo2")
		assert.True(t, ok)
		assert.Equal(t, "bar2", val)
	}

	c.Add("foo1", "bar11", 1)
	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, "bar11", val)
	}
}

func TestHeapCache_Get(t *testing.T) {
	c := NewHeapCache(10)

	c.Add("foo1", "bar1", 1)
	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, "bar1", val)
	}
	{
		val, ok := c.Get("foo2")
		assert.False(t, ok)
		assert.Nil(t, val)
	}
}

func TestHeapCache_Len(t *testing.T) {
	c := NewHeapCache(10)

	c.Add("foo1", "bar1", 1)

	assert.Equal(t, 1, c.Len())
}

func TestHeapCache_AddMany(t *testing.T) {
	c := NewHeapCache(3)

	item1 := HeapCacheItem{Key: "foo1", Value: "bar1", Priority: 1}
	item2 := HeapCacheItem{Key: "foo2", Value: "bar2", Priority: 2}
	item3 := HeapCacheItem{Key: "foo3", Value: "bar3", Priority: 3}
	item4 := HeapCacheItem{Key: "foo4", Value: "bar4", Priority: 4}

	c.AddMany(item1, item2)

	assert.Equal(t, 2, c.Len())

	assert.True(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))

	item1.Priority = 100
	c.AddMany(item1, item3, item4)

	assert.Equal(t, 3, c.Len())

	assert.False(t, c.Contains("foo2"))
	assert.True(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo3"))
	assert.True(t, c.Contains("foo4"))
}

func TestHeapCache_Evict(t *testing.T) {
	var i int
	capacity := 50
	n := 100

	c := NewHeapCache(uint(capacity))

	for i = 0; i < n; i++ {
		c.Add(i, i, int64(i))
	}

	assert.Equal(t, int(math.Min(float64(capacity), float64(n))), c.Len())

	for i = 0; i < n; i++ {
		if i < n-capacity {
			assert.False(t, c.Contains(i))
		} else {
			assert.True(t, c.Contains(i))
		}
	}
}

func TestHeapCache_Remove(t *testing.T) {
	c := NewHeapCache(10)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 1)

	assert.True(t, c.Remove("foo1"))
	assert.False(t, c.Remove("foo1"))

	assert.False(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))
}

func TestHeapCache_Contains(t *testing.T) {
	c := NewHeapCache(10)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 1)

	assert.True(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))

	c.Remove("foo1")

	assert.False(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))
}

func TestHeapCache_Priority(t *testing.T) {
	c := NewHeapCache(3)

	c.Add("foo1", "bar1", 10)
	c.Add("foo2", "bar2", 20)
	c.Add("foo3", "bar3", 30)

	assert.True(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))
	assert.True(t, c.Contains("foo3"))

	c.Add("foo4", "bar4", 40)
	assert.Equal(t, c.Len(), 3)
	assert.True(t, c.Contains("foo4"))
	assert.False(t, c.Contains("foo1"))

	c.Add("foo3", "bar3", 10)
	assert.Equal(t, c.Len(), 3)
	assert.True(t, c.Contains("foo3"))

	c.Add("foo5", "bar5", 40)
	assert.Equal(t, c.Len(), 3)
	assert.True(t, c.Contains("foo5"))
	assert.False(t, c.Contains("foo3"))
}

func TestItemsHeap_ZeroCapacity(t *testing.T) {
	c := NewHeapCache(0)

	c.Add("foo", "bar", 1)
	c.AddMany(HeapCacheItem{Key: "foo", Value: "bar", Priority: 1})
	assert.False(t, c.Contains("foo"))
}

func BenchmarkHeapCache_Add(b *testing.B) {
	c := NewHeapCache(uint(b.N))

	for n := 0; n < b.N; n++ {
		c.Add(n, n, int64(n))
	}
}

func BenchmarkHeapCache_AddWithEvictHalf(b *testing.B) {
	c := NewHeapCache(uint(b.N / 2))

	for n := 0; n < b.N; n++ {
		c.Add(n, n, int64(n))
	}
}

func BenchmarkHeapCache_Get(b *testing.B) {
	c := NewHeapCache(uint(b.N))

	for n := 0; n < b.N; n++ {
		c.Get(n)
	}
}
