// Package identity is the use case of signing in: create an account, start
// and end sessions, and tell who a session token belongs to.
package identity

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
)

// ErrInvalidCredentials is the one answer to a wrong email or a wrong
// password, so the answer never tells which accounts exist.
var ErrInvalidCredentials = fmt.Errorf("invalid credentials: %w", domain.ErrUnauthenticated)

// ThrottledError is returned when one address has failed to sign in too often.
type ThrottledError struct {
	RetryAfter time.Duration
}

func (e *ThrottledError) Error() string {
	return fmt.Sprintf("too many sign-in attempts, retry in %s", e.RetryAfter.Round(time.Second))
}

// Accounts keeps accounts.
type Accounts interface {
	// Create stores an account, and returns account.ErrEmailTaken when another
	// one already uses the address.
	Create(ctx context.Context, a account.Account) error
	// ByEmail returns the account with a normalised address, or domain.ErrNotFound.
	ByEmail(ctx context.Context, email string) (account.Account, error)
}

// Sessions keeps sessions, found by the hash of their token.
type Sessions interface {
	Create(ctx context.Context, s account.Session) error
	// Find returns a session and its account, or domain.ErrNotFound.
	Find(ctx context.Context, hash account.TokenHash) (account.Session, account.Account, error)
	Extend(ctx context.Context, hash account.TokenHash, expiresAt time.Time) error
	Delete(ctx context.Context, hash account.TokenHash) error
	// DeleteExpired removes every session over by now and says how many.
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

// PasswordHasher turns passwords into slow, salted hashes and checks them.
type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password, encoded string) (bool, error)
}

// Limiter counts sign-in attempts per key.
type Limiter interface {
	// Allow spends one attempt for key, or says how long until one is free.
	Allow(key string) (allowed bool, retryAfter time.Duration)
	// Reset forgets the attempts of key, after it signed in.
	Reset(key string)
}

// Deps is what a Service needs. Every field is required.
type Deps struct {
	Accounts Accounts
	Sessions Sessions
	Hasher   PasswordHasher
	Limiter  Limiter
	Now      func() time.Time
}

// Service runs the use case.
type Service struct {
	accounts Accounts
	sessions Sessions
	hasher   PasswordHasher
	limiter  Limiter
	now      func() time.Time
	// decoy is checked against when no account has the address, so a missing
	// account takes as long to refuse as a wrong password.
	decoy string
}

// New returns a Service. It hashes a decoy password first, which takes as
// long as one sign-in.
func New(ctx context.Context, deps Deps) (*Service, error) {
	decoy, err := deps.Hasher.Hash(ctx, rand.Text())
	if err != nil {
		return nil, fmt.Errorf("hash the decoy password: %w", err)
	}
	return &Service{
		accounts: deps.Accounts,
		sessions: deps.Sessions,
		hasher:   deps.Hasher,
		limiter:  deps.Limiter,
		now:      deps.Now,
		decoy:    decoy,
	}, nil
}

// Grant is a session that just started. Token reaches the caller once and is
// stored nowhere: the database keeps its hash.
type Grant struct {
	Token   string
	Session account.Session
	Account account.Account
}

// Principal is who a request acts for.
type Principal struct {
	AccountID uuid.UUID
	Email     string
	ExpiresAt time.Time
	TokenHash account.TokenHash
}

// SignUp creates an account and starts its first session.
func (s *Service) SignUp(ctx context.Context, email, password string) (Grant, error) {
	var vs domain.Violations
	normalized, code, ok := account.NormalizeEmail(email)
	if !ok {
		vs.Add(account.FieldEmail, code)
	}
	if code, ok := account.CheckPassword(password); !ok {
		vs.Add(account.FieldPassword, code)
	}
	if err := vs.Err(); err != nil {
		return Grant{}, err
	}

	hash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return Grant{}, fmt.Errorf("hash password: %w", err)
	}
	acc := account.Account{
		ID:           uuid.NewV7(),
		Email:        normalized,
		PasswordHash: hash,
		CreatedAt:    s.now().UTC().Truncate(time.Microsecond),
	}
	if err := s.accounts.Create(ctx, acc); err != nil {
		return Grant{}, fmt.Errorf("create account: %w", err)
	}
	return s.start(ctx, acc)
}

// SignIn checks an email and a password and starts a session.
func (s *Service) SignIn(ctx context.Context, email, password string) (Grant, error) {
	normalized, _, ok := account.NormalizeEmail(email)
	key := normalized
	if !ok {
		key = strings.ToLower(strings.TrimSpace(email))
	}
	if allowed, wait := s.limiter.Allow(key); !allowed {
		return Grant{}, &ThrottledError{RetryAfter: wait}
	}

	acc, err := account.Account{}, domain.ErrNotFound
	if ok {
		acc, err = s.accounts.ByEmail(ctx, normalized)
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		if _, err := s.hasher.Verify(ctx, password, s.decoy); err != nil {
			return Grant{}, fmt.Errorf("verify the decoy password: %w", err)
		}
		return Grant{}, ErrInvalidCredentials
	case err != nil:
		return Grant{}, fmt.Errorf("find account: %w", err)
	}

	match, err := s.hasher.Verify(ctx, password, acc.PasswordHash)
	if err != nil {
		return Grant{}, fmt.Errorf("verify password: %w", err)
	}
	if !match {
		return Grant{}, ErrInvalidCredentials
	}
	s.limiter.Reset(key)
	return s.start(ctx, acc)
}

// Authenticate returns who a session token belongs to, pushing the session's
// deadline forward when it is in the second half of its life.
func (s *Service) Authenticate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, domain.ErrUnauthenticated
	}
	hash := account.HashToken(token)
	sess, acc, err := s.sessions.Find(ctx, hash)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return Principal{}, domain.ErrUnauthenticated
	case err != nil:
		return Principal{}, fmt.Errorf("find session: %w", err)
	}

	now := s.now()
	if sess.Expired(now) {
		return Principal{}, domain.ErrUnauthenticated
	}
	if sess.DueForRenewal(now) {
		sess.ExpiresAt = now.UTC().Truncate(time.Microsecond).Add(account.SessionTTL)
		if err := s.sessions.Extend(ctx, hash, sess.ExpiresAt); err != nil {
			return Principal{}, fmt.Errorf("extend session: %w", err)
		}
	}
	return Principal{AccountID: acc.ID, Email: acc.Email, ExpiresAt: sess.ExpiresAt, TokenHash: hash}, nil
}

// SignOut ends the session p acts through.
func (s *Service) SignOut(ctx context.Context, p Principal) error {
	if err := s.sessions.Delete(ctx, p.TokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// PurgeExpiredSessions deletes the sessions that are over and says how many.
func (s *Service) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	n, err := s.sessions.DeleteExpired(ctx, s.now())
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return n, nil
}

func (s *Service) start(ctx context.Context, acc account.Account) (Grant, error) {
	token := rand.Text() // 128 random bits
	sess := account.NewSession(account.HashToken(token), acc.ID, s.now())
	if err := s.sessions.Create(ctx, sess); err != nil {
		return Grant{}, fmt.Errorf("create session: %w", err)
	}
	return Grant{Token: token, Session: sess, Account: acc}, nil
}
