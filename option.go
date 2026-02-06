package memoize

import "time"

type config struct {
	ttl time.Duration
}

// Option configures memoization behavior.
type Option func(*config)

// WithTTL sets the cache expiration duration. Default is no expiration.
func WithTTL(d time.Duration) Option {
	return func(c *config) {
		c.ttl = d
	}
}
