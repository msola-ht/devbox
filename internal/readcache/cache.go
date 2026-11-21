// Package readcache retains the last successful read within one app process.
package readcache

import (
	"context"
	"sync"
	"time"
)

// Cache serializes reads of one value. Its zero value is ready to use.
type Cache[T any] struct {
	init   sync.Once
	slot   chan struct{}
	value  T
	readAt time.Time
}

// Read reuses the cached value unless refresh is requested or maxAge has passed.
// A zero maxAge keeps the value for the lifetime of the cache.
// The loader returns an error if its read was interrupted.
func (c *Cache[T]) Read(ctx context.Context, refresh bool, maxAge time.Duration, load func(context.Context) (T, error)) (T, error) {
	c.init.Do(func() { c.slot = make(chan struct{}, 1) })
	var zero T
	select {
	case c.slot <- struct{}{}:
		defer func() { <-c.slot }()
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if !refresh && !c.readAt.IsZero() && (maxAge == 0 || time.Since(c.readAt) < maxAge) {
		return c.value, nil
	}
	value, err := load(ctx)
	if err != nil {
		return zero, err
	}
	c.value = value
	c.readAt = time.Now()
	return value, nil
}
