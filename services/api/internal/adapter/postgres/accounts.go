package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres/sqlcdb"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
)

// Accounts stores accounts. It implements identity.Accounts.
type Accounts struct {
	q *sqlcdb.Queries
}

// NewAccounts returns Accounts on pool.
func NewAccounts(pool *pgxpool.Pool) *Accounts {
	return &Accounts{q: sqlcdb.New(pool)}
}

// Create stores a, or returns account.ErrEmailTaken.
func (s *Accounts) Create(ctx context.Context, a account.Account) error {
	err := s.q.CreateAccount(ctx, sqlcdb.CreateAccountParams{
		ID:           a.ID,
		Email:        a.Email,
		PasswordHash: a.PasswordHash,
		CreatedAt:    a.CreatedAt,
	})
	switch {
	case isUniqueViolation(err, "accounts_email_key"):
		return account.ErrEmailTaken
	case err != nil:
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

// ByEmail returns the account with a normalised address, or domain.ErrNotFound.
func (s *Accounts) ByEmail(ctx context.Context, email string) (account.Account, error) {
	row, err := s.q.AccountByEmail(ctx, email)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return account.Account{}, domain.ErrNotFound
	case err != nil:
		return account.Account{}, fmt.Errorf("select account: %w", err)
	}
	return account.Account{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash, CreatedAt: row.CreatedAt.UTC()}, nil
}

// Sessions stores sessions. It implements identity.Sessions.
type Sessions struct {
	q *sqlcdb.Queries
}

// NewSessions returns Sessions on pool.
func NewSessions(pool *pgxpool.Pool) *Sessions {
	return &Sessions{q: sqlcdb.New(pool)}
}

// Create stores a session.
func (s *Sessions) Create(ctx context.Context, sess account.Session) error {
	err := s.q.CreateSession(ctx, sqlcdb.CreateSessionParams{
		TokenHash: sess.TokenHash[:],
		AccountID: sess.AccountID,
		CreatedAt: sess.CreatedAt,
		ExpiresAt: sess.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Find returns the session with the token hash and its account, or
// domain.ErrNotFound. The account comes without its password hash, which no
// caller of Find needs.
func (s *Sessions) Find(ctx context.Context, hash account.TokenHash) (account.Session, account.Account, error) {
	row, err := s.q.FindSession(ctx, hash[:])
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return account.Session{}, account.Account{}, domain.ErrNotFound
	case err != nil:
		return account.Session{}, account.Account{}, fmt.Errorf("select session: %w", err)
	}
	sess := account.Session{TokenHash: hash, AccountID: row.AccountID, CreatedAt: row.CreatedAt.UTC(), ExpiresAt: row.ExpiresAt.UTC()}
	acc := account.Account{ID: row.AccountID, Email: row.Email, CreatedAt: row.AccountCreatedAt.UTC()}
	return sess, acc, nil
}

// Extend moves a session's deadline.
func (s *Sessions) Extend(ctx context.Context, hash account.TokenHash, expiresAt time.Time) error {
	if err := s.q.ExtendSession(ctx, sqlcdb.ExtendSessionParams{TokenHash: hash[:], ExpiresAt: expiresAt}); err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// Delete removes a session. Removing one that is already gone is not an error:
// signing out twice ends in the same place.
func (s *Sessions) Delete(ctx context.Context, hash account.TokenHash) error {
	if err := s.q.DeleteSession(ctx, hash[:]); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpired removes the sessions over by now and says how many.
func (s *Sessions) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	n, err := s.q.DeleteExpiredSessions(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return n, nil
}
