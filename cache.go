package cache

import (
    "fmt"
    "sync"
    "time"
)

type Item[T any] struct {
    Object     T
    Expiration int64
}

func (item Item[T]) Expire() bool {
    if item.Expiration == 0 {
        return false
    }
    return time.Now().UnixNano() > item.Expiration
}

const (
    NoExpiration      time.Duration = -1
    DefaultExpiration time.Duration = 0
)

type Cache[T any] struct {
    *cache[T]
}

type cache[T any] struct {
    defaultExpiration time.Duration
    items             map[string]Item[T]
    mu                sync.RWMutex
    onEvicted         func(string, any)
}

func (c *cache[T]) Set(k string, x T, d time.Duration) {
    var e int64
    if d == DefaultExpiration {
        d = c.defaultExpiration
    }
    if d > 0 {
        e = time.Now().Add(d).UnixNano()
    }
    c.mu.Lock()
    c.items[k] = Item[T]{
        Object:     x,
        Expiration: e,
    }
    c.mu.Unlock()
}

func (c *cache[T]) set(k string, x T, d time.Duration) {
    var e int64
    if d == DefaultExpiration {
        d = c.defaultExpiration
    }
    if d > 0 {
        e = time.Now().Add(d).UnixNano()
    }
    c.items[k] = Item[T]{
        Object:     x,
        Expiration: e,
    }
}

func (c *cache[T]) SetDefault(k string, x T) {
    c.Set(k, x, DefaultExpiration)
}

func (c *cache[T]) Add(k string, x T, d time.Duration) error {
    c.mu.Lock()
    _, found := c.get(k)
    if found {
        c.mu.Unlock()
        return fmt.Errorf("Item %s already exists", k)
    }
    c.set(k, x, d)
    c.mu.Unlock()
    return nil
}

func (c *cache[T]) Replace(k string, x T, d time.Duration) error {
    c.mu.Lock()
    _, found := c.get(k)
    if !found {
        c.mu.Unlock()
        return fmt.Errorf("Item %s doesn't exist", k)
    }
    c.set(k, x, d)
    c.mu.Unlock()
    return nil
}

func (c *cache[T]) Get(k string) (any, bool) {
    c.mu.RLock()
    item, found := c.items[k]
    if !found {
        c.mu.RUnlock()
        return nil, false
    }
    if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
        c.mu.RUnlock()
        return nil, false
    }
    c.mu.RUnlock()
    return item.Object, true
}

func (c *cache[T]) get(k string) (value T, ok bool) {
    item, found := c.items[k]
    if !found {
        return value, false
    }
    if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
        return value, false
    }
    return item.Object, true
}

func newCache[T any](de time.Duration, m map[string]Item[T]) *cache[T] {
    if de == DefaultExpiration {
        de = NoExpiration
    }
    c := &cache[T]{
        defaultExpiration: de,
        items:             m,
    }
    return c
}

func newCacheWithJanitor[T any](de time.Duration, ci time.Duration, m map[string]Item[T]) *Cache[T] {
    c := newCache(de, m)
    // This trick ensures that the janitor goroutine (which--granted it
    // was enabled--is running DeleteExpired on c forever) does not keep
    // the returned C object from being garbage collected. When it is
    // garbage collected, the finalizer stops the janitor goroutine, after
    // which c can be collected.
    C := &Cache[T]{c}
    if ci > 0 {
        // pass
    }
    return C
}

func New[T any](defaultExpiration, cleanupInterval time.Duration) *Cache[T] {
    items := make(map[string]Item[T])
    return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}
