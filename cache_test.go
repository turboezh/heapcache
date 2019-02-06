package heapcache

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cacheItem struct {
	Value    string
	Priority int
}

func testLess(a, b interface{}) bool {
	return a.(*cacheItem).Priority < b.(*cacheItem).Priority
}

func TestCache_Add(t *testing.T) {
	c := New(10, testLess)

	foo1 := &cacheItem{"bar1", 1}
	foo2 := &cacheItem{"bar2", 2}
	c.Add("foo1", foo1)
	c.Add("foo2", foo2)

	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, foo1, val)
	}
	{
		val, ok := c.Get("foo2")
		assert.True(t, ok)
		assert.Equal(t, foo2, val)
	}

	foo1 = &cacheItem{"bar123", 1}
	c.Add("foo1", foo1)
	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, "bar123", val.(*cacheItem).Value)
	}
}

func TestCache_Get(t *testing.T) {
	c := New(10, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	{
		val, ok := c.Get("foo1")
		assert.True(t, ok)
		assert.Equal(t, "bar1", val.(*cacheItem).Value)
	}
	{
		val, ok := c.Get("foo2")
		assert.False(t, ok)
		assert.Nil(t, val)
	}
}

func TestCache_Len(t *testing.T) {
	c := New(10, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})

	assert.Equal(t, 1, c.Len())
}

func TestCache_evict(t *testing.T) {
	var i int
	capacity := 50
	n := 100

	c := New(capacity, testLess)

	for i = 0; i < n; i++ {
		v := strconv.FormatInt(int64(i), 10)
		c.Add(i, &cacheItem{v, i})
	}

	assert.Equal(t, int(math.Min(float64(capacity), float64(n))), c.Len())

	for i = 0; i < n; i++ {
		if i < n-capacity {
			assert.False(t, c.All(i))
		} else {
			assert.True(t, c.All(i))
		}
	}
}

func TestCache_Remove(t *testing.T) {
	c := New(10, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 2})
	c.Add("foo3", &cacheItem{"bar3", 3})

	assert.Equal(t, 1, c.Remove("foo1"))
	assert.Equal(t, 0, c.Remove("foo1"))
	assert.Equal(t, 2, c.Len())

	assert.False(t, c.All("foo1"))
	assert.True(t, c.All("foo2"))

	assert.Equal(t, 2, c.Remove("foo1", "foo2", "foo3"))
	assert.Equal(t, 0, c.Len())
}

func TestCache_All(t *testing.T) {
	c := New(10, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 1})

	assert.False(t, c.All("foo0"))
	assert.True(t, c.All("foo1"))
	assert.True(t, c.All("foo1", "foo2"))
	assert.False(t, c.All("foo1", "foo2", "foo3"))
}

func TestCache_Any(t *testing.T) {
	c := New(10, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 1})

	assert.False(t, c.Any("foo0"))
	assert.True(t, c.Any("foo1"))
	assert.True(t, c.Any("foo1", "foo2"))
	assert.True(t, c.Any("foo1", "foo2", "foo3"))
	assert.False(t, c.Any("foo4", "foo5", "foo6"))
}

func TestCache_Priority(t *testing.T) {
	c := New(3, testLess)

	c.Add("foo1", &cacheItem{"bar1", 10})
	c.Add("foo2", &cacheItem{"bar2", 20})
	c.Add("foo3", &cacheItem{"bar3", 30})

	assert.True(t, c.All("foo1"))
	assert.True(t, c.All("foo2"))
	assert.True(t, c.All("foo3"))

	c.Add("foo4", &cacheItem{"bar4", 40})
	assert.Equal(t, 3, c.Len())
	assert.True(t, c.All("foo4"))
	assert.False(t, c.All("foo1"))

	c.Add("foo5", &cacheItem{"bar5", 10})
	assert.Equal(t, 3, c.Len())
	assert.True(t, c.All("foo5"))
	assert.False(t, c.All("foo2"))

	c.Add("foo6", &cacheItem{"bar6", 40})
	assert.Equal(t, 3, c.Len())
	assert.True(t, c.All("foo6"))
	assert.False(t, c.All("foo5"))
}

func TestCache_ZeroCapacity(t *testing.T) {
	c := New(0, testLess)

	c.Add("foo", &cacheItem{"bar", 1})
	c.Add("foo", &cacheItem{"bar", 1})
	assert.False(t, c.All("foo"))
}

func TestCache_Purge(t *testing.T) {
	c := New(3, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 1})

	assert.Equal(t, 2, c.Len())

	c.Purge()

	assert.Equal(t, 0, c.Len())
}

func TestCache_Evict(t *testing.T) {
	c := New(3, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 2})
	c.Add("foo3", &cacheItem{"bar3", 3})

	assert.Equal(t, c.Len(), 3)

	evicted := c.Evict(2)
	assert.Equal(t, 2, evicted)
	assert.Equal(t, 1, c.Len())

	// overflow
	evicted = c.Evict(2)
	assert.Equal(t, 1, evicted)
	assert.Equal(t, 0, c.Len())

	evicted = c.Evict(2)
	assert.Equal(t, 0, evicted)
	assert.Equal(t, 0, c.Len())

	evicted = c.Evict(0)
	assert.Equal(t, 0, evicted)
	assert.Equal(t, 0, c.Len())
}

func TestCache_Capacity(t *testing.T) {
	c := New(3, testLess)
	assert.Equal(t, 3, c.Capacity())
}

func TestCache_ChangeCapacity(t *testing.T) {
	c := New(3, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 2})
	c.Add("foo3", &cacheItem{"bar3", 3})

	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 3, c.Capacity())

	// noop
	c.ChangeCapacity(0)
	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 3, c.Capacity())

	assert.True(t, c.All("foo1", "foo2", "foo3"))

	// expand
	c.ChangeCapacity(2)
	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 5, c.Capacity())

	assert.True(t, c.All("foo1", "foo2", "foo3"))

	// shrink
	c.ChangeCapacity(-2)
	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 3, c.Capacity())

	assert.True(t, c.All("foo1", "foo2", "foo3"))

	// shrink with evict
	c.ChangeCapacity(-2)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 1, c.Capacity())

	assert.True(t, c.All("foo3"))
	assert.False(t, c.All("foo1", "foo2"))

	c.ChangeCapacity(2)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 3, c.Capacity())

	assert.True(t, c.All("foo3"))
	assert.False(t, c.All("foo1", "foo2"))
}

func TestCache_SetCapacityUnderflow(t *testing.T) {
	c := New(3, testLess)
	c.SetCapacity(-5)
	assert.Equal(t, 0, c.Capacity())
}

func TestCache_ChangeCapacityUnderflow(t *testing.T) {
	c := New(3, testLess)
	c.ChangeCapacity(-5)
	assert.Equal(t, 0, c.Capacity())
}

func TestCache_SetCapacity(t *testing.T) {
	c := New(3, testLess)

	c.Add("foo1", &cacheItem{"bar1", 1})
	c.Add("foo2", &cacheItem{"bar2", 2})
	c.Add("foo3", &cacheItem{"bar3", 3})

	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 3, c.Capacity())

	// expand
	c.SetCapacity(5)
	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 5, c.Capacity())

	assert.True(t, c.All("foo1", "foo2", "foo3"))

	// shrink
	c.SetCapacity(3)
	assert.Equal(t, 3, c.Len())
	assert.Equal(t, 3, c.Capacity())

	assert.True(t, c.All("foo1", "foo2", "foo3"))

	// shrink with evict
	c.SetCapacity(1)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 1, c.Capacity())

	assert.True(t, c.All("foo3"))
	assert.False(t, c.All("foo1", "foo2"))

	c.SetCapacity(3)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 3, c.Capacity())

	assert.True(t, c.All("foo3"))
	assert.False(t, c.All("foo1", "foo2"))
}

func BenchmarkCache_Add(b *testing.B) {
	c := New(b.N, testLess)

	item := &cacheItem{"", 0}
	for n := 0; n < b.N; n++ {
		c.Add(n, item)
	}
}

func BenchmarkCache_AddWithEvictHalf(b *testing.B) {
	c := New(b.N/2, testLess)

	item := &cacheItem{"", 0}
	for n := 0; n < b.N; n++ {
		c.Add(n, item)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := New(b.N, testLess)

	for n := 0; n < b.N; n++ {
		c.Get(n)
	}
}

func Example() {
	type Foo struct {
		Value     int
		Timestamp time.Time
	}

	cache := New(10, func(a, b interface{}) bool {
		return a.(*Foo).Timestamp.Before(b.(*Foo).Timestamp)
	})

	item1 := Foo{10, time.Now()}
	item2 := Foo{20, time.Now().Add(time.Second)}

	cache.Add("one", &item1)
	cache.Add("two", &item2)
}
