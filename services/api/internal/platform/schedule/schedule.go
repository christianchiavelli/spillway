// Package schedule runs housekeeping on a timer inside the server process.
package schedule

import (
	"context"
	"time"
)

// Every calls fn each interval until ctx is done, and then returns nil. A run
// still going when ctx ends gets the cancelled ctx and should stop early. fn
// reports its own errors: one failed run must not stop the next.
func Every(ctx context.Context, interval time.Duration, fn func(context.Context)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fn(ctx)
		}
	}
}
