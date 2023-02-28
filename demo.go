package cache

type demoCache[T any] struct {
    items map[string]T
}

func (d *demoCache[T]) Set(k string, item T) {
    d.items[k] = item
}

func (d *demoCache[T]) Get(k string) T {
    return d.items[k]
}
