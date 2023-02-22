package cache

import (
	"sync"
	"time"
)

const (
	NoExpiration         time.Duration = -1
	DefaultPurgeInterval               = time.Minute
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]interface{}
	wd    *watchdog
}

func New(purgeInterval time.Duration) *Cache {
	c := Cache{
		mu:    sync.RWMutex{},
		items: map[string]interface{}{},
	}
	c.wd = newWatchdog(&c, purgeInterval)
	c.wd.run()
	return &c
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	val, ok := c.items[key]
	c.mu.RUnlock()
	return val, ok
}

func (c *Cache) Set(key string, value interface{}, lifespan time.Duration) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
	if lifespan != NoExpiration {
		c.wd.track(key, time.Now().Add(lifespan).UnixNano())
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	c.wd.untrack(key)
}

func (c *Cache) ItemCount() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

func (c *Cache) Flush() {
	c.mu.Lock()
	c.items = map[string]interface{}{}
	c.mu.Unlock()
}

type watchdog struct {
	cache         *Cache
	expirationSet map[string]int64
	mu            sync.Mutex
	purgeInterval time.Duration
}

func newWatchdog(cache *Cache, purgeInterval time.Duration) *watchdog {
	wd := watchdog{
		cache:         cache,
		expirationSet: map[string]int64{},
		mu:            sync.Mutex{},
		purgeInterval: purgeInterval,
	}
	return &wd
}

func (w *watchdog) run() {
	purgeTimer := time.NewTicker(w.purgeInterval)
	go w.delete(purgeTimer.C)
}

func (w *watchdog) track(key string, expiresAt int64) {
	w.mu.Lock()
	w.expirationSet[key] = expiresAt
	w.mu.Unlock()
}

func (w *watchdog) untrack(key string) {
	w.mu.Lock()
	delete(w.expirationSet, key)
	w.mu.Unlock()
}

func (w *watchdog) delete(tickCh <-chan time.Time) {
	for {
		select {
		case <-tickCh:
			now := time.Now().UnixNano()
			evictedKeys := make([]string, 0, len(w.expirationSet))
			w.mu.Lock()
			for key, expiresAt := range w.expirationSet {
				if now > expiresAt {
					w.cache.mu.Lock()
					delete(w.cache.items, key)
					w.cache.mu.Unlock()
					evictedKeys = append(evictedKeys, key)
				}
			}
			for i := range evictedKeys {
				delete(w.expirationSet, evictedKeys[i])
			}
			w.mu.Unlock()
		default:
			continue
		}
	}
}
