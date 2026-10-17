package argon2id

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAHashWaitingForASlotGivesUpWithItsRequest(t *testing.T) {
	h := New(Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}, 1)
	h.slots <- struct{}{} // another hash holds the only slot

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if _, err := h.Hash(ctx, "waits its turn"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Hash() = %v, want context.DeadlineExceeded", err)
	}

	<-h.slots
	if _, err := h.Hash(t.Context(), "runs once the slot is free"); err != nil {
		t.Fatalf("Hash() after the slot was freed = %v", err)
	}
}
