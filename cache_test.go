package heapcache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Add(t *testing.T) {
	c := New(10)

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

func TestCache_Get(t *testing.T) {
	c := New(10)

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

func TestCache_Len(t *testing.T) {
	c := New(10)

	c.Add("foo1", "bar1", 1)

	assert.Equal(t, 1, c.Len())
}

func TestCache_AddMany(t *testing.T) {
	c := New(3)

	item1 := Item{Key: "foo1", Value: "bar1", Priority: 1}
	item2 := Item{Key: "foo2", Value: "bar2", Priority: 2}
	item3 := Item{Key: "foo3", Value: "bar3", Priority: 3}
	item4 := Item{Key: "foo4", Value: "bar4", Priority: 4}

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

func TestCache_evict(t *testing.T) {
	var i int
	capacity := 50
	n := 100

	c := New(uint(capacity))

	for i = 0; i < n; i++ {
		c.Add(i, i, PriorityType(i))
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

func TestCache_Remove(t *testing.T) {
	c := New(10)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 2)
	c.Add("foo3", "bar3", 3)

	assert.Equal(t, 1, c.Remove("foo1"))
	assert.Equal(t, 0, c.Remove("foo1"))
	assert.Equal(t, 2, c.Len())

	assert.False(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))

	assert.Equal(t, 2, c.Remove("foo1", "foo2", "foo3"))
	assert.Equal(t, 0, c.Len())
}

func TestCache_Contains(t *testing.T) {
	c := New(10)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 1)

	assert.True(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))

	c.Remove("foo1")

	assert.False(t, c.Contains("foo1"))
	assert.True(t, c.Contains("foo2"))
}

func TestCache_Priority(t *testing.T) {
	c := New(3)

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

func TestCache_ZeroCapacity(t *testing.T) {
	c := New(0)

	c.Add("foo", "bar", 1)
	c.AddMany(Item{Key: "foo", Value: "bar", Priority: 1})
	assert.False(t, c.Contains("foo"))
}

func TestCache_Purge(t *testing.T) {
	c := New(3)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 1)

	assert.Equal(t, c.Len(), 2)

	c.Purge()

	assert.Equal(t, c.Len(), 0)
}

func TestCache_Evict(t *testing.T) {
	c := New(3)

	c.Add("foo1", "bar1", 1)
	c.Add("foo2", "bar2", 2)
	c.Add("foo3", "bar3", 3)

	assert.Equal(t, c.Len(), 3)

	evicted := c.Evict(2)
	assert.Equal(t, evicted, 2)
	assert.Equal(t, c.Len(), 1)

	// overflow
	evicted = c.Evict(2)
	assert.Equal(t, evicted, 1)
	assert.Equal(t, c.Len(), 0)

	evicted = c.Evict(2)
	assert.Equal(t, evicted, 0)
	assert.Equal(t, c.Len(), 0)

	evicted = c.Evict(0)
	assert.Equal(t, evicted, 0)
	assert.Equal(t, c.Len(), 0)
}

func BenchmarkCache_Add(b *testing.B) {
	c := New(uint(b.N))

	for n := 0; n < b.N; n++ {
		c.Add(n, n, PriorityType(n))
	}
}

func BenchmarkCache_AddWithEvictHalf(b *testing.B) {
	c := New(uint(b.N / 2))

	for n := 0; n < b.N; n++ {
		c.Add(n, n, PriorityType(n))
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := New(uint(b.N))

	for n := 0; n < b.N; n++ {
		c.Get(n)
	}
}
