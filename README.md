# heapcache
[![Build Status](https://travis-ci.org/turboezh/heapcache.svg)](https://travis-ci.org/turboezh/heapcache)
[![GitHub release](https://img.shields.io/github/release/turboezh/heapcache.svg)](https://github.com/turboezh/heapcache/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/turboezh/heapcache)](https://goreportcard.com/report/github.com/turboezh/heapcache)
[![Maintainability](https://api.codeclimate.com/v1/badges/de484103003b548529f0/maintainability)](https://codeclimate.com/github/turboezh/heapcache/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/de484103003b548529f0/test_coverage)](https://codeclimate.com/github/turboezh/heapcache/test_coverage)
![Downloads](https://img.shields.io/github/downloads/turboezh/heapcahce/total.svg)
[![GoDoc](https://godoc.org/github.com/turboezh/heapcache?status.svg)](https://godoc.org/github.com/turboezh/heapcache)

This cache implementation is based on priority queue (see [Heap](https://golang.org/pkg/container/heap/)).
It uses user-defined comparator to evaluate priorities of cached items. Items with lowest priorities will be evicted first.

Features:
 - simple standard data structure;
 - no write locks on get operations;
 - capacity may be changed at any time.

# Requirements
Go >= 1.11

# Documentation
https://godoc.org/github.com/turboezh/heapcache

# Examples

## Cache
```go
type Foo struct {
    Value int
    Timestamp time.Time
}

item1 := Foo{10, time.Now()}
item2 := Foo{20, time.Now().Add(time.Second)}

cache := New(10, func(a, b interface{}) bool {
    return a.(*Foo).Timestamp.Before(b.(*Foo).Timestamp)
})
```

## Add item
```go
cache.Add("one", &item1)
cache.Add("two", &item2)

```

## Get item
```go
item, exists := cache.Get("one")
if !exists {
    // `foo` doesn't exists in cache
    // `item` is nil
}
// cache returns `interface{}` so we need to assert type (if need so)
item = item.(*Foo) // = &item1
```

## Check item
```go
// check if cache contain all keys 
ok := cache.All("one", "two")

// check if cache contain any of keys 
ok := cache.Any("one", "two")
```

## Remove item
```go
// Remove returns false if there is no item in cache
wasRemoved := cache.Remove("one")
```

## Support on Beerpay
Hey dude! Help me out for a couple of :beers:!

[![Beerpay](https://beerpay.io/turboezh/heapcache/badge.svg?style=beer-square)](https://beerpay.io/turboezh/heapcache)  [![Beerpay](https://beerpay.io/turboezh/heapcache/make-wish.svg?style=flat-square)](https://beerpay.io/turboezh/heapcache?focus=wish)
