package memoize

import "sync"

type call[V any] struct {
	wg  sync.WaitGroup
	val V
	err error
}

type singleflight[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]*call[V]
}

// do execute fn once per key, deduplicating concurrent calls with the same key.
// All callers with the same key receive the same result.
func (sf *singleflight[K, V]) do(key K, fn func() (V, error)) (V, error) {
	sf.mu.Lock()
	if sf.m == nil {
		sf.m = make(map[K]*call[V])
	}
	if c, ok := sf.m[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &call[V]{}
	c.wg.Add(1)
	sf.m[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.val, c.err
}
