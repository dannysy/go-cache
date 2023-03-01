# go-cache

go-cache is an in-memory key/value store. It is a thread-safe cache with expiration times.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines. go-cache has a watcher with its own set of tracked items,
so it locks cache only on eviction of expired items.

### Installation

`go get github.com/dannysy/go-cache`

### Usage

```go
package main

import (
	"fmt"
	"time"
	
	"github.com/dannysy/go-cache"
)

func main() {
	// Create a cache which purges expired items every 5 minutes
	c := cache.New(5*time.Minute)

	// Set the value of the key "foo" to "bar", with the 15 minutes expiration time
	c.Set("foo", "bar", 15*time.Minute)

	// Set the value of the key "bar" to "foo", with no expiration time
	c.Set("bar", "foo", cache.NoExpiration)
	
	// Res-et the value of the key "bar" to "foo1", with no expiration time
	c.Set("bar", "foo1", cache.NoExpiration)
	
	// Delete item by the key "bar"
	c.Delete("bar")

	// Get the value of the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Printf("as interface{} %v\n", foo)
		fmt.Printf("as string %v\n", foo.(string))
	}
	
	//flush all cache items
	c.Flush()
}
```
