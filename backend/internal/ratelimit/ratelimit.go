// Package ratelimit implements in-memory keyed token buckets with periodic cleanup.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

// Limiter allows N events per window per key (token bucket with capacity N).
type Limiter struct {
	mu       sync.Mutex
	capacity float64
	perToken time.Duration
	buckets  map[string]*bucket
	now      func() time.Time
}

// New returns a limiter allowing n events per window per key.
func New(n int, window time.Duration) *Limiter {
	if n < 1 {
		n = 1
	}
	return &Limiter{
		capacity: float64(n),
		perToken: window / time.Duration(n),
		buckets:  make(map[string]*bucket),
		now:      time.Now,
	}
}

func (l *Limiter) refill(b *bucket, now time.Time) {
	if el := now.Sub(b.last); el > 0 {
		b.tokens = min(l.capacity, b.tokens+float64(el)/float64(l.perToken))
		b.last = now
	}
}

// Allow consumes a token for key. When denied it returns how long until a token is available.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, last: now}
		l.buckets[key] = b
	}
	l.refill(b, now)
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	wait := time.Duration((1 - b.tokens) * float64(l.perToken))
	if wait < time.Second {
		wait = time.Second
	}
	return false, wait
}

// Cleanup removes buckets that have fully refilled (equivalent to absent).
func (l *Limiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	for k, b := range l.buckets {
		l.refill(b, now)
		if b.tokens >= l.capacity {
			delete(l.buckets, k)
		}
	}
}

// Len returns the number of tracked keys.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

// StartCleanup runs Cleanup every interval until ctx is done.
func (l *Limiter) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				l.Cleanup()
			}
		}
	}()
}
