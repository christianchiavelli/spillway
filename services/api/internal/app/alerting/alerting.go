// Package alerting is the use case of keeping alerts: create, list and delete
// the alerts of a signed-in account.
package alerting

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
)

// Store keeps alerts. It is declared here, by the code that needs it, and holds
// only what this use case asks of storage.
type Store interface {
	// Add stores a, unless its account already holds max alerts, which returns
	// alert.ErrLimitReached, or one by the same name, which returns
	// alert.ErrNameTaken. Two calls for one account must not both pass the
	// limit check, so the count and the insert are atomic per account.
	Add(ctx context.Context, a alert.Alert, max int) error
	// ListByAccount returns the account's alerts, newest first.
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]alert.Alert, error)
	// Delete removes one of the account's alerts, and returns domain.ErrNotFound
	// when the account has no alert by that id.
	Delete(ctx context.Context, accountID, alertID uuid.UUID) error
}

// Service runs the use case.
type Service struct {
	store Store
	now   func() time.Time
}

// New returns a Service that reads the time from now.
func New(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

// Create checks a draft and stores the alert it describes for the account.
func (s *Service) Create(ctx context.Context, accountID uuid.UUID, d alert.Draft) (alert.Alert, error) {
	a, err := alert.New(uuid.NewV7(), accountID, d, s.now())
	if err != nil {
		return alert.Alert{}, err
	}
	if err := s.store.Add(ctx, a, alert.MaxPerAccount); err != nil {
		return alert.Alert{}, fmt.Errorf("add alert: %w", err)
	}
	return a, nil
}

// List returns the account's alerts, newest first.
func (s *Service) List(ctx context.Context, accountID uuid.UUID) ([]alert.Alert, error) {
	alerts, err := s.store.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return alerts, nil
}

// Delete removes one of the account's alerts. Someone else's alert is
// reported as not found, so ids cannot be probed.
func (s *Service) Delete(ctx context.Context, accountID, alertID uuid.UUID) error {
	if err := s.store.Delete(ctx, accountID, alertID); err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	return nil
}
