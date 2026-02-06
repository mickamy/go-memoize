package memoize

import (
	"sync"
	"time"
)

type entry[V any] struct {
	val       V
	expiresAt time.Time
}

func (e *entry[V]) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

// Memo provides a memoized function with cache management capabilities.
type Memo[K comparable, V any] struct {
	fn    func(K) (V, error)
	cfg   config
	mu    sync.RWMutex
	cache map[K]*entry[V]
	sf    singleflight[K, V]
}

// New creates a new memoized function wrapper.
func New[K comparable, V any](fn func(K) (V, error), opts ...Option) *Memo[K, V] {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	return &Memo[K, V]{
		fn:    fn,
		cfg:   cfg,
		cache: make(map[K]*entry[V]),
	}
}

// Get executes the function or returns a cached result.
func (m *Memo[K, V]) Get(key K) (V, error) {
	m.mu.RLock()
	if e, ok := m.cache[key]; ok && !e.expired() {
		m.mu.RUnlock()
		return e.val, nil
	}
	m.mu.RUnlock()

	return m.sf.do(key, func() (V, error) {
		// Double-check after acquiring singleflight.
		m.mu.RLock()
		if e, ok := m.cache[key]; ok && !e.expired() {
			m.mu.RUnlock()
			return e.val, nil
		}
		m.mu.RUnlock()

		val, err := m.fn(key)
		if err != nil {
			return val, err
		}

		e := &entry[V]{val: val}
		if m.cfg.ttl > 0 {
			e.expiresAt = time.Now().Add(m.cfg.ttl)
		}

		m.mu.Lock()
		m.cache[key] = e
		m.mu.Unlock()

		return val, nil
	})
}

// Forget removes the cached result for the given key.
func (m *Memo[K, V]) Forget(key K) {
	m.mu.Lock()
	delete(m.cache, key)
	m.mu.Unlock()
}

// Purge removes all cached results.
func (m *Memo[K, V]) Purge() {
	m.mu.Lock()
	m.cache = make(map[K]*entry[V])
	m.mu.Unlock()
}

// Do wrap a function with memoization and returns the wrapped function.
// For simple cases where cache management is not needed.
func Do[K comparable, V any](fn func(K) (V, error), opts ...Option) func(K) (V, error) {
	m := New(fn, opts...)
	return m.Get
}
