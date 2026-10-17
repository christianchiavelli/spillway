package schedule_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/schedule"
)

func TestEveryRunsOnTheIntervalUntilCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		runs := 0
		done := make(chan error)
		go func() {
			done <- schedule.Every(ctx, time.Minute, func(context.Context) { runs++ })
		}()

		time.Sleep(3*time.Minute + time.Second)
		synctest.Wait()
		if runs != 3 {
			t.Errorf("after 3 minutes: %d runs, want 3", runs)
		}

		cancel()
		if err := <-done; err != nil {
			t.Errorf("Every() = %v, want nil", err)
		}
	})
}
