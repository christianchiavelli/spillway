// Package alerting is the use case of keeping alerts: create, list and delete
// the alerts of a signed-in account, and read the days they fired.
package alerting

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
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
	// Activity returns, for each of the account's alerts that ever fired, the
	// last day it did up to through, and on how many days from from it did.
	Activity(ctx context.Context, accountID uuid.UUID, from, through time.Time) (map[uuid.UUID]alert.Activity, error)
	// Firings returns a page of the days from from to through on which the
	// account's alerts fired, newest first, and how many there are in all.
	Firings(ctx context.Context, accountID uuid.UUID, from, through time.Time, limit, offset int) ([]alert.Firing, int, error)
}

// Calendar says how far the ONS's data goes.
type Calendar interface {
	// Bounds returns curtailment.ErrNoData until the first file is loaded.
	Bounds(ctx context.Context) (curtailment.Bounds, error)
}

// Watched is an alert, with what its threshold met in the data.
type Watched struct {
	alert.Alert
	alert.Activity
}

// FiringPage is a page of firings in the days an alert's activity covers.
// From and Through are zero, and the page empty, before any data is loaded.
type FiringPage struct {
	From, Through time.Time
	Total         int
	Items         []alert.Firing
}

// Service runs the use case.
type Service struct {
	store    Store
	calendar Calendar
	now      func() time.Time
}

// New returns a Service that reads the time from now.
func New(store Store, calendar Calendar, now func() time.Time) *Service {
	return &Service{store: store, calendar: calendar, now: now}
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

// List returns the account's alerts, newest first, each with its activity.
func (s *Service) List(ctx context.Context, accountID uuid.UUID) ([]Watched, error) {
	alerts, err := s.store.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	activity, err := s.activity(ctx, accountID)
	if err != nil {
		return nil, err
	}
	out := make([]Watched, len(alerts))
	for i, a := range alerts {
		out[i] = Watched{Alert: a, Activity: activity[a.ID]}
	}
	return out, nil
}

// Delete removes one of the account's alerts. Someone else's alert is
// reported as not found, so ids cannot be probed.
func (s *Service) Delete(ctx context.Context, accountID, alertID uuid.UUID) error {
	if err := s.store.Delete(ctx, accountID, alertID); err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	return nil
}

// Firings returns a page of the days the account's alerts fired in the
// alert.RecentDays to the latest published, newest first.
func (s *Service) Firings(ctx context.Context, accountID uuid.UUID, limit, offset int) (FiringPage, error) {
	from, through, ok, err := s.window(ctx)
	if err != nil || !ok {
		return FiringPage{Items: []alert.Firing{}}, err
	}
	items, total, err := s.store.Firings(ctx, accountID, from, through, limit, offset)
	if err != nil {
		return FiringPage{}, fmt.Errorf("list firings: %w", err)
	}
	return FiringPage{From: from, Through: through, Total: total, Items: items}, nil
}

func (s *Service) activity(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]alert.Activity, error) {
	from, through, ok, err := s.window(ctx)
	if err != nil || !ok {
		return nil, err
	}
	activity, err := s.store.Activity(ctx, accountID, from, through)
	if err != nil {
		return nil, fmt.Errorf("read activity: %w", err)
	}
	return activity, nil
}

// window is the alert.RecentDays up to the latest day published. ok is false
// before any data is loaded: an alert then has no activity, not an error.
func (s *Service) window(ctx context.Context) (from, through time.Time, ok bool, err error) {
	b, err := s.calendar.Bounds(ctx)
	switch {
	case errors.Is(err, curtailment.ErrNoData):
		return time.Time{}, time.Time{}, false, nil
	case err != nil:
		return time.Time{}, time.Time{}, false, fmt.Errorf("read the bounds: %w", err)
	}
	through = b.LastDay()
	return through.AddDate(0, 0, 1-alert.RecentDays), through, true, nil
}
