package ratelimit_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/ratelimit"
)

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func TestAllowSpendsTheBurstThenWaits(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
	k := ratelimit.New(12*time.Second, 5, c.now)

	for i := range 5 {
		if ok, _ := k.Allow("ana@example.com"); !ok {
			t.Fatalf("attempt %d refused", i+1)
		}
	}
	ok, wait := k.Allow("ana@example.com")
	if ok || wait != 12*time.Second {
		t.Fatalf("attempt 6 = %v, %v; want refused for 12s", ok, wait)
	}
	if ok, _ := k.Allow("bia@example.com"); !ok {
		t.Error("another key was refused")
	}

	c.advance(12 * time.Second)
	if ok, _ := k.Allow("ana@example.com"); !ok {
		t.Error("refused after a refill")
	}
}

func TestARefusedAttemptCostsNothing(t *testing.T) {
	c := &clock{t: time.Now()}
	k := ratelimit.New(time.Minute, 1, c.now)
	k.Allow("key")

	for range 10 {
		k.Allow("key") // refused, and must not push the wait further
	}
	if _, wait := k.Allow("key"); wait != time.Minute {
		t.Errorf("wait = %v, want one refill", wait)
	}
}

func TestResetAndSweepForgetKeys(t *testing.T) {
	c := &clock{t: time.Now()}
	k := ratelimit.New(10*time.Second, 2, c.now)
	k.Allow("signed in")
	k.Allow("guessing")
	k.Allow("guessing")

	k.Reset("signed in")
	if k.Len() != 1 {
		t.Fatalf("Len() after Reset() = %d, want 1", k.Len())
	}

	c.advance(10 * time.Second)
	if n := k.Sweep(); n != 0 {
		t.Errorf("Sweep() forgot %d keys with half a bucket", n)
	}
	c.advance(10 * time.Second)
	if n := k.Sweep(); n != 1 || k.Len() != 0 {
		t.Errorf("Sweep() = %d, Len() = %d; want the full bucket forgotten", n, k.Len())
	}
}

func TestAllowIsSafeForConcurrentUse(t *testing.T) {
	k := ratelimit.New(time.Hour, 5, time.Now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			if ok, _ := k.Allow(fmt.Sprint("key", i%4)); ok {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if got := allowed.Load(); got != 4*5 {
		t.Errorf("allowed %d attempts, want 5 for each of 4 keys", got)
	}
}
