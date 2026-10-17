package identity_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/app/identity"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
)

const password = "correct horse battery staple"

func TestSignUpStartsASession(t *testing.T) {
	f := newFixture(t)

	grant, err := f.svc.SignUp(t.Context(), " Ana@Example.com ", password)
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}

	if grant.Account.Email != "ana@example.com" {
		t.Errorf("email = %q, want it normalised", grant.Account.Email)
	}
	if grant.Account.PasswordHash == password || grant.Account.PasswordHash == "" {
		t.Errorf("password hash = %q", grant.Account.PasswordHash)
	}
	if len(grant.Token) < 26 {
		t.Errorf("token %q is shorter than 128 bits of base32", grant.Token)
	}
	stored, ok := f.sessions.byHash[account.HashToken(grant.Token)]
	if !ok || stored != grant.Session {
		t.Errorf("stored session = %+v, want %+v", stored, grant.Session)
	}
}

func TestSignUpReportsEveryBrokenRule(t *testing.T) {
	f := newFixture(t)

	_, err := f.svc.SignUp(t.Context(), "not an address", "short")

	var verr *domain.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("SignUp() error = %v, want a *domain.ValidationError", err)
	}
	want := []domain.Violation{
		{Field: "email", Code: domain.CodeInvalid},
		{Field: "password", Code: domain.CodeTooShort},
	}
	if !slices.Equal(verr.Violations, want) {
		t.Errorf("violations = %v, want %v", verr.Violations, want)
	}
}

func TestSignUpRefusesAnAddressInUse(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.SignUp(t.Context(), "ana@example.com", password); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.SignUp(t.Context(), "ANA@example.com", password)
	if !errors.Is(err, account.ErrEmailTaken) || !errors.Is(err, domain.ErrConflict) {
		t.Errorf("SignUp() error = %v, want account.ErrEmailTaken", err)
	}
}

func TestSignInAnswersTheSameForAnyWrongCredential(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.SignUp(t.Context(), "ana@example.com", password); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct{ email, password string }{
		{"ana@example.com", "wrong horse battery staple"},
		{"bia@example.com", password},
		{"not an address", password},
	} {
		before := f.hasher.verified
		_, err := f.svc.SignIn(t.Context(), tt.email, tt.password)
		if !errors.Is(err, identity.ErrInvalidCredentials) || !errors.Is(err, domain.ErrUnauthenticated) {
			t.Errorf("SignIn(%q) error = %v, want identity.ErrInvalidCredentials", tt.email, err)
		}
		if f.hasher.verified != before+1 {
			t.Errorf("SignIn(%q) checked %d hashes, want 1 whether or not the account exists", tt.email, f.hasher.verified-before)
		}
	}
}

func TestSignInIsThrottledPerAddress(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.SignUp(t.Context(), "ana@example.com", password); err != nil {
		t.Fatal(err)
	}
	f.limiter.retryAfter = 30 * time.Second

	_, err := f.svc.SignIn(t.Context(), "Ana@Example.com", password)

	var throttled *identity.ThrottledError
	if !errors.As(err, &throttled) || throttled.RetryAfter != 30*time.Second {
		t.Fatalf("SignIn() error = %v, want a ThrottledError of 30s", err)
	}
	if got := f.limiter.allowed; !slices.Equal(got, []string{"ana@example.com"}) {
		t.Errorf("limiter keys = %q, want the normalised address", got)
	}

	f.limiter.retryAfter = 0
	if _, err := f.svc.SignIn(t.Context(), "ana@example.com", password); err != nil {
		t.Fatalf("SignIn() once allowed again: %v", err)
	}
	if !slices.Equal(f.limiter.reset, []string{"ana@example.com"}) {
		t.Errorf("limiter resets = %q, want one after signing in", f.limiter.reset)
	}
}

func TestAuthenticate(t *testing.T) {
	f := newFixture(t)
	grant, err := f.svc.SignUp(t.Context(), "ana@example.com", password)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a fresh session keeps its deadline", func(t *testing.T) {
		f.clock.set(grant.Session.CreatedAt.Add(time.Hour))
		p, err := f.svc.Authenticate(t.Context(), grant.Token)
		if err != nil {
			t.Fatal(err)
		}
		if p.AccountID != grant.Account.ID || p.Email != "ana@example.com" || !p.ExpiresAt.Equal(grant.Session.ExpiresAt) {
			t.Errorf("Authenticate() = %+v", p)
		}
		if f.sessions.extended != 0 {
			t.Errorf("extended %d times", f.sessions.extended)
		}
	})

	t.Run("an older session moves its deadline forward", func(t *testing.T) {
		now := grant.Session.CreatedAt.Add(5 * 24 * time.Hour)
		f.clock.set(now)
		p, err := f.svc.Authenticate(t.Context(), grant.Token)
		if err != nil {
			t.Fatal(err)
		}
		if want := now.Add(account.SessionTTL); !p.ExpiresAt.Equal(want) {
			t.Errorf("ExpiresAt = %v, want %v", p.ExpiresAt, want)
		}
		if f.sessions.extended != 1 {
			t.Errorf("extended %d times, want 1", f.sessions.extended)
		}
	})

	t.Run("an expired session is refused", func(t *testing.T) {
		f.clock.set(f.sessions.byHash[account.HashToken(grant.Token)].ExpiresAt)
		if _, err := f.svc.Authenticate(t.Context(), grant.Token); !errors.Is(err, domain.ErrUnauthenticated) {
			t.Errorf("Authenticate() error = %v, want domain.ErrUnauthenticated", err)
		}
	})

	t.Run("an unknown or empty token is refused", func(t *testing.T) {
		for _, token := range []string{"", "nope"} {
			if _, err := f.svc.Authenticate(t.Context(), token); !errors.Is(err, domain.ErrUnauthenticated) {
				t.Errorf("Authenticate(%q) error = %v, want domain.ErrUnauthenticated", token, err)
			}
		}
	})
}

func TestSignOutEndsTheSession(t *testing.T) {
	f := newFixture(t)
	grant, err := f.svc.SignUp(t.Context(), "ana@example.com", password)
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.svc.Authenticate(t.Context(), grant.Token)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.svc.SignOut(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Authenticate(t.Context(), grant.Token); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Errorf("Authenticate() after SignOut() = %v, want domain.ErrUnauthenticated", err)
	}
}

func TestPurgeExpiredSessions(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.SignUp(t.Context(), "ana@example.com", password); err != nil {
		t.Fatal(err)
	}
	f.clock.set(f.clock.now().Add(account.SessionTTL))

	n, err := f.svc.PurgeExpiredSessions(t.Context())
	if err != nil || n != 1 || len(f.sessions.byHash) != 0 {
		t.Errorf("PurgeExpiredSessions() = %d, %v; %d left", n, err, len(f.sessions.byHash))
	}
}

type fixture struct {
	svc      *identity.Service
	sessions *memSessions
	hasher   *plainHasher
	limiter  *fakeLimiter
	clock    *clock
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		sessions: &memSessions{byHash: map[account.TokenHash]account.Session{}},
		hasher:   &plainHasher{},
		limiter:  &fakeLimiter{},
		clock:    &clock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)},
	}
	accounts := &memAccounts{}
	f.sessions.accounts = accounts
	svc, err := identity.New(t.Context(), identity.Deps{
		Accounts: accounts,
		Sessions: f.sessions,
		Hasher:   f.hasher,
		Limiter:  f.limiter,
		Now:      f.clock.now,
	})
	if err != nil {
		t.Fatal(err)
	}
	f.svc = svc
	return f
}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = t
}

type memAccounts struct {
	list []account.Account
}

func (m *memAccounts) Create(_ context.Context, a account.Account) error {
	for _, b := range m.list {
		if b.Email == a.Email {
			return account.ErrEmailTaken
		}
	}
	m.list = append(m.list, a)
	return nil
}

func (m *memAccounts) ByEmail(_ context.Context, email string) (account.Account, error) {
	for _, a := range m.list {
		if a.Email == email {
			return a, nil
		}
	}
	return account.Account{}, domain.ErrNotFound
}

type memSessions struct {
	accounts *memAccounts
	byHash   map[account.TokenHash]account.Session
	extended int
}

func (m *memSessions) Create(_ context.Context, s account.Session) error {
	m.byHash[s.TokenHash] = s
	return nil
}

func (m *memSessions) Find(_ context.Context, hash account.TokenHash) (account.Session, account.Account, error) {
	s, ok := m.byHash[hash]
	if !ok {
		return account.Session{}, account.Account{}, domain.ErrNotFound
	}
	for _, a := range m.accounts.list {
		if a.ID == s.AccountID {
			return s, a, nil
		}
	}
	return account.Session{}, account.Account{}, domain.ErrNotFound
}

func (m *memSessions) Extend(_ context.Context, hash account.TokenHash, expiresAt time.Time) error {
	s := m.byHash[hash]
	s.ExpiresAt = expiresAt
	m.byHash[hash] = s
	m.extended++
	return nil
}

func (m *memSessions) Delete(_ context.Context, hash account.TokenHash) error {
	delete(m.byHash, hash)
	return nil
}

func (m *memSessions) DeleteExpired(_ context.Context, now time.Time) (int64, error) {
	var n int64
	for h, s := range m.byHash {
		if s.Expired(now) {
			delete(m.byHash, h)
			n++
		}
	}
	return n, nil
}

// plainHasher stands in for argon2id, which takes a tenth of a second a call.
type plainHasher struct {
	verified int
}

func (h *plainHasher) Hash(_ context.Context, password string) (string, error) {
	return "plain$" + password, nil
}

func (h *plainHasher) Verify(_ context.Context, password, encoded string) (bool, error) {
	h.verified++
	return strings.TrimPrefix(encoded, "plain$") == password, nil
}

type fakeLimiter struct {
	retryAfter time.Duration
	allowed    []string
	reset      []string
}

func (l *fakeLimiter) Allow(key string) (bool, time.Duration) {
	l.allowed = append(l.allowed, key)
	return l.retryAfter == 0, l.retryAfter
}

func (l *fakeLimiter) Reset(key string) {
	l.reset = append(l.reset, key)
}
