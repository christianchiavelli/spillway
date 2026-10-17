package postgres

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres/sqlcdb"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
)

// Alerts stores alerts. It implements alerting.Store.
type Alerts struct {
	pool *pgxpool.Pool
	q    *sqlcdb.Queries
}

// NewAlerts returns Alerts on pool.
func NewAlerts(pool *pgxpool.Pool) *Alerts {
	return &Alerts{pool: pool, q: sqlcdb.New(pool)}
}

// Add stores a unless its account holds max alerts already, or one by the
// same name.
//
// The limit takes three statements in one transaction: lock the account's row,
// count, insert. The lock makes a second Add for the same account wait until
// the first commits, and under READ COMMITTED each statement reads what was
// committed before it began, so the second count sees the first insert. A
// single INSERT ... WHERE (SELECT count(*) ...) < max would not do: its count
// is read before it waits, and both inserts would pass.
func (s *Alerts) Add(ctx context.Context, a alert.Alert, max int) error {
	return pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := q.LockAccount(ctx, a.AccountID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("lock account: %w", domain.ErrNotFound)
			}
			return fmt.Errorf("lock account: %w", err)
		}
		held, err := q.CountAlerts(ctx, a.AccountID)
		if err != nil {
			return fmt.Errorf("count alerts: %w", err)
		}
		if held >= int64(max) {
			return alert.ErrLimitReached
		}
		err = q.InsertAlert(ctx, sqlcdb.InsertAlertParams{
			ID:           a.ID,
			AccountID:    a.AccountID,
			Name:         a.Name,
			Source:       string(a.Source),
			Subsystem:    string(a.Subsystem),
			ThresholdMwh: a.ThresholdMWh,
			CreatedAt:    a.CreatedAt,
		})
		switch {
		case isUniqueViolation(err, "alerts_account_name_key"):
			return alert.ErrNameTaken
		case err != nil:
			return fmt.Errorf("insert alert: %w", err)
		}
		return nil
	})
}

// ListByAccount returns the account's alerts, newest first.
func (s *Alerts) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]alert.Alert, error) {
	rows, err := s.q.ListAlerts(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("select alerts: %w", err)
	}
	alerts := make([]alert.Alert, len(rows))
	for i, r := range rows {
		alerts[i] = alert.Alert{
			ID:           r.ID,
			AccountID:    r.AccountID,
			Name:         r.Name,
			Source:       alert.Source(r.Source),
			Subsystem:    alert.Subsystem(r.Subsystem),
			ThresholdMWh: r.ThresholdMwh,
			CreatedAt:    r.CreatedAt.UTC(),
		}
	}
	return alerts, nil
}

// Delete removes one of the account's alerts, or returns domain.ErrNotFound.
func (s *Alerts) Delete(ctx context.Context, accountID, alertID uuid.UUID) error {
	n, err := s.q.DeleteAlert(ctx, sqlcdb.DeleteAlertParams{ID: alertID, AccountID: accountID})
	switch {
	case err != nil:
		return fmt.Errorf("delete alert: %w", err)
	case n == 0:
		return domain.ErrNotFound
	}
	return nil
}
