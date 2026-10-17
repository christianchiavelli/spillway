// Package account holds the rules of who can sign in: what an email address
// and a password must look like, and how long a session lives.
package account

import (
	"crypto/sha256"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
)

// The limits the API contract repeats.
const (
	EmailMaxLength    = 254 // RFC 5321's limit on a path, less the angle brackets
	PasswordMinLength = 15  // NIST SP 800-63B-4, for a password that is the only factor
	PasswordMaxLength = 128 // long enough for any passphrase, short enough to hash cheaply
)

// Field names, as the API contract spells them.
const (
	FieldEmail    = "email"
	FieldPassword = "password"
)

// ErrEmailTaken is returned when an account already uses the address.
var ErrEmailTaken error = &domain.ConflictError{Field: FieldEmail, Code: domain.CodeTaken}

// Account is someone who can sign in. PasswordHash never leaves the service.
type Account struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// NormalizeEmail returns the form an address is stored and compared in, or
// the code of the rule it breaks. Case is folded because every mail provider
// people use ignores it, though RFC 5321 allows the local part to keep it.
func NormalizeEmail(s string) (string, domain.Code, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case s == "":
		return "", domain.CodeRequired, false
	case len(s) > EmailMaxLength:
		return "", domain.CodeTooLong, false
	}
	addr, err := mail.ParseAddress(s)
	// ParseAddress also accepts `Name <a@b>`; only a bare address is one.
	if err != nil || addr.Address != s || !strings.Contains(s[strings.LastIndexByte(s, '@')+1:], ".") {
		return "", domain.CodeInvalid, false
	}
	return s, "", true
}

// CheckPassword returns the code of the rule a new password breaks, if any.
// There are no rules about which characters to use: NIST found they make
// passwords weaker, not stronger.
func CheckPassword(p string) (domain.Code, bool) {
	switch n := utf8.RuneCountInString(p); {
	case n == 0:
		return domain.CodeRequired, false
	case n < PasswordMinLength:
		return domain.CodeTooShort, false
	case n > PasswordMaxLength:
		return domain.CodeTooLong, false
	}
	return "", true
}

// SessionTTL is how long a session lasts without use. Using it in the second
// half of its life moves the deadline forward again.
const SessionTTL = 7 * 24 * time.Hour

// Session is a signed-in browser. Only the hash of its token is stored: a leak
// of the table hands out no way in.
type Session struct {
	TokenHash TokenHash
	AccountID uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}

// TokenHash is the SHA-256 of a session token. A token carries 128 random
// bits, so a fast hash is enough; a slow one would only slow every request.
type TokenHash [sha256.Size]byte

// HashToken returns the hash a session is stored and looked up by.
func HashToken(token string) TokenHash {
	return sha256.Sum256([]byte(token))
}

// NewSession starts a session for an account at now.
func NewSession(hash TokenHash, accountID uuid.UUID, now time.Time) Session {
	now = now.UTC().Truncate(time.Microsecond)
	return Session{TokenHash: hash, AccountID: accountID, CreatedAt: now, ExpiresAt: now.Add(SessionTTL)}
}

// Expired reports whether the session is over at now.
func (s Session) Expired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

// DueForRenewal reports whether using the session at now should push its
// deadline forward. Waiting for the second half of its life keeps that to
// about one write per session every few days, not one per request.
func (s Session) DueForRenewal(now time.Time) bool {
	return s.ExpiresAt.Sub(now) < SessionTTL/2
}
