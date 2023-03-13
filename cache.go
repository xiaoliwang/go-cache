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

func (item Item[T]) Expired() bool {
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
	if item.Expired() {
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()
	return item.Object, true
}

func (c *cache[T]) GetWithExpiration(k string) (v T, t time.Time, ok bool) {
	c.mu.RLock()
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return v, t, ok
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			c.mu.RUnlock()
			return v, t, ok
		}
		c.mu.RUnlock()
		return item.Object, time.Unix(0, item.Expiration), true
	}
	c.mu.RUnlock()
	return item.Object, t, true
}

func (c *cache[T]) get(k string) (v T, ok bool) {
	item, found := c.items[k]
	if !found {
		return v, ok
	}
	if item.Expired() {
		return v, ok
	}
	return item.Object, true
}

func (c *cache[T]) Increment(k string, n int64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}

	switch vo := any(v.Object).(type) {
	case int:
		v.Object = any(vo + int(n)).(T)
	case int8:
		v.Object = any(vo + int8(n)).(T)
	case int16:
		v.Object = any(vo + int16(n)).(T)
	case int32:
		v.Object = any(vo + int32(n)).(T)
	case int64:
		v.Object = any(vo + n).(T)
	case uint:
		v.Object = any(vo + uint(n)).(T)
	case uintptr:
		v.Object = any(vo + uintptr(n)).(T)
	case uint8:
		v.Object = any(vo + uint8(n)).(T)
	case uint16:
		v.Object = any(vo + uint16(n)).(T)
	case uint32:
		v.Object = any(vo + uint32(n)).(T)
	case uint64:
		v.Object = any(vo + uint64(n)).(T)
	case float32:
		v.Object = any(vo + float32(n)).(T)
	case float64:
		v.Object = any(vo + float64(n)).(T)
	default:
		c.mu.Unlock()
		return fmt.Errorf("The value for %s is not an integer", k)
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

func (c *cache[T]) IncrementFloat(k string, n float64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch vo := any(v.Object).(type) {
	case float32:
		v.Object = any(vo + float32(n)).(T)
	case float64:
		v.Object = any(vo + n).(T)
	default:
		c.mu.Unlock()
		return fmt.Errorf("The value for %s does not have type float32 or float64", k)
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

func (c *cache[T]) Decrement(k string, n int64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}

	switch vo := any(v.Object).(type) {
	case int:
		v.Object = any(vo - int(n)).(T)
	case int8:
		v.Object = any(vo - int8(n)).(T)
	case int16:
		v.Object = any(vo - int16(n)).(T)
	case int32:
		v.Object = any(vo - int32(n)).(T)
	case int64:
		v.Object = any(vo - n).(T)
	case uint:
		v.Object = any(vo - uint(n)).(T)
	case uintptr:
		v.Object = any(vo - uintptr(n)).(T)
	case uint8:
		v.Object = any(vo - uint8(n)).(T)
	case uint16:
		v.Object = any(vo - uint16(n)).(T)
	case uint32:
		v.Object = any(vo - uint32(n)).(T)
	case uint64:
		v.Object = any(vo - uint64(n)).(T)
	case float32:
		v.Object = any(vo - float32(n)).(T)
	case float64:
		v.Object = any(vo - float64(n)).(T)
	default:
		c.mu.Unlock()
		return fmt.Errorf("The value for %s is not an integer", k)
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

func (c *cache[T]) DecrementFloat(k string, n float64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch vo := any(v.Object).(type) {
	case float32:
		v.Object = any(vo - float32(n)).(T)
	case float64:
		v.Object = any(vo - n).(T)
	default:
		c.mu.Unlock()
		return fmt.Errorf("The value for %s does not have type float32 or float64", k)
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

func (c *cache[T]) Delete(k string) {
	c.mu.Lock()
	v, evicted := c.delete(k)
	c.mu.Unlock()
	if evicted {
		c.onEvicted(k, v)
	}
}

func (c *cache[T]) delete(k string) (vo T, ok bool) {
	if c.onEvicted != nil {
		if v, found := c.items[k]; found {
			delete(c.items, k)
			return v.Object, true
		}
	}
	delete(c.items, k)
	return vo, ok
}

type keyAndValue[T any] struct {
	key   string
	value T
}

func (c *cache[T]) DeleteExpired() {
	var evictedItems []keyAndValue[any]
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			ov, evicted := c.delete(k)
			if evicted {
				evictedItems = append(evictedItems, keyAndValue[any]{k, ov})
			}
		}
	}
	c.mu.Unlock()
	for _, v := range evictedItems {
		c.onEvicted(v.key, v.value)
	}
}

func (c *cache[T]) OnEvicted(f func(string, any)) {
	c.mu.Lock()
	c.onEvicted = f
	c.mu.Unlock()
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
