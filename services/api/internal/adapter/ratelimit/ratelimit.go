// Package ratelimit keeps a token bucket per key in memory. Spillway uses it to
// slow down password guessing against one address.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Keyed holds one bucket per key. A bucket holds burst attempts and gets one
// back every refill. It lives in one process, so with several replicas each
// one counts on its own; ADR 0007 explains why that is enough here.
type Keyed struct {
	mu      sync.Mutex
	buckets map[string]*rate.Limiter
	refill  time.Duration
	burst   int
	now     func() time.Time
}

// New returns a Keyed that reads the time from now.
func New(refill time.Duration, burst int, now func() time.Time) *Keyed {
	return &Keyed{buckets: map[string]*rate.Limiter{}, refill: refill, burst: burst, now: now}
}

// Allow spends one attempt for key, or says how long until one is free.
func (k *Keyed) Allow(key string) (bool, time.Duration) {
	now := k.now()
	k.mu.Lock()
	defer k.mu.Unlock()

	lim, ok := k.buckets[key]
	if !ok {
		lim = rate.NewLimiter(rate.Every(k.refill), k.burst)
		k.buckets[key] = lim
	}
	r := lim.ReserveN(now, 1)
	if wait := r.DelayFrom(now); wait > 0 {
		r.CancelAt(now) // a refused attempt costs nothing
		return false, wait
	}
	return true, 0
}

// Reset forgets key, after it signed in.
func (k *Keyed) Reset(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.buckets, key)
}

// Sweep forgets every key whose bucket is full again, since a full bucket is
// what a new key starts with, and says how many it forgot. Without it, a
// stream of made-up addresses would grow the map for as long as it lasts.
func (k *Keyed) Sweep() int {
	now := k.now()
	k.mu.Lock()
	defer k.mu.Unlock()

	n := 0
	for key, lim := range k.buckets {
		if lim.TokensAt(now) >= float64(k.burst) {
			delete(k.buckets, key)
			n++
		}
	}
	return n
}

// Len says how many keys are held.
func (k *Keyed) Len() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.buckets)
}
