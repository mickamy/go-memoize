package memoize

// NewSingleflight creates a new singleflight instance for testing.
func NewSingleflight[K comparable, V any]() *singleflight[K, V] {
	return &singleflight[K, V]{}
}

// DoSingleflight calls the unexported do method for testing.
func (sf *singleflight[K, V]) Do(key K, fn func() (V, error)) (V, error) {
	return sf.do(key, fn)
}
