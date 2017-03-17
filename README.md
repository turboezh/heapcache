# heapcache
[![Build Status](https://travis-ci.org/turboezh/heapcache.svg)](https://travis-ci.org/turboezh/heapcache)
[![codecov](https://codecov.io/gh/turboezh/heapcache/branch/master/graph/badge.svg)](https://codecov.io/gh/turboezh/heapcache)
[![Go Report Card](https://goreportcard.com/badge/github.com/turboezh/heapcache)](https://goreportcard.com/report/github.com/turboezh/heapcache)
[![GoDoc](https://godoc.org/github.com/turboezh/heapcache?status.svg)](https://godoc.org/github.com/turboezh/heapcache)

Heap based cache with 'priority' evict policy

# Installation
`go get github.com/turboezh/heapcache`


# Documentation
https://godoc.org/github.com/turboezh/heapcache


# Examples

```go
cache := heapcache.New(3)


// add value to cache
// `key` and `value` may be any `interface{}` not just string (see https://golang.org/ref/spec#KeyType)
cache.Add("foo1", "bar1", 1)
cache.Add("foo2", "bar2", 2)

value, ok := cache.Get("foo1")
if !ok {
    // `foo1` doesn't exists in cache
    // `value` is nil
}
// cache operates with `interface{}` so we need to assert type (if need so)
valueString := value.(string)


// just check existence
isExists := cache.Contains("foo1")


// add many items at once (it's optimized)
foo3 := heapcache.Item{Key: "foo3", Value: "bar3", Priority: 3}
foo4 := heapcache.Item{Key: "foo4", Value: "bar4", Priority: 4}

// "foo1" will be evicted as it has lowest priority
cache.AddMany(foo3, foo4)
cacheLen := cache.Len() // == 3 (foo2, foo3, foo4)


// Add will renew item's `value` and `priority`
cache.Add("foo2", "bar222", 100)


// Remove returns false if there is no item in cache
wasRemoved := cache.Remove("foo3")
```
