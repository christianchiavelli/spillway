package account_test

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		in       string
		want     string
		wantCode domain.Code
	}{
		{in: "  Ana.Souza@Example.COM ", want: "ana.souza@example.com"},
		{in: "", wantCode: domain.CodeRequired},
		{in: "   ", wantCode: domain.CodeRequired},
		{in: "ana", wantCode: domain.CodeInvalid},
		{in: "ana@localhost", wantCode: domain.CodeInvalid},
		{in: "Ana <ana@example.com>", wantCode: domain.CodeInvalid},
		{in: "ana@@example.com", wantCode: domain.CodeInvalid},
		{in: strings.Repeat("a", 243) + "@example.com", wantCode: domain.CodeTooLong},
	}
	for _, tt := range tests {
		got, code, ok := account.NormalizeEmail(tt.in)
		if ok != (tt.wantCode == "") || got != tt.want || code != tt.wantCode {
			t.Errorf("NormalizeEmail(%q) = %q, %q, %v; want %q, %q", tt.in, got, code, ok, tt.want, tt.wantCode)
		}
	}
}

func TestCheckPasswordCountsCharactersNotBytes(t *testing.T) {
	tests := []struct {
		in       string
		wantCode domain.Code
	}{
		{in: "", wantCode: domain.CodeRequired},
		{in: strings.Repeat("a", 14), wantCode: domain.CodeTooShort},
		{in: strings.Repeat("a", 15)},
		{in: strings.Repeat("ç", 15)}, // 30 bytes, 15 characters
		{in: "correct horse battery staple"},
		{in: strings.Repeat("a", 128)},
		{in: strings.Repeat("a", 129), wantCode: domain.CodeTooLong},
	}
	for _, tt := range tests {
		code, ok := account.CheckPassword(tt.in)
		if ok != (tt.wantCode == "") || code != tt.wantCode {
			t.Errorf("CheckPassword(%d characters) = %q, %v; want %q", len([]rune(tt.in)), code, ok, tt.wantCode)
		}
	}
}

func TestSessionLifetime(t *testing.T) {
	start := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	s := account.NewSession(account.HashToken("token"), uuid.NewV7(), start)

	if got, want := s.ExpiresAt, start.Add(7*24*time.Hour); !got.Equal(want) {
		t.Fatalf("ExpiresAt = %v, want %v", got, want)
	}

	tests := []struct {
		at          time.Duration
		expired     bool
		renewalDue  bool
		description string
	}{
		{at: 0, description: "just started"},
		{at: 84 * time.Hour, description: "exactly half way"},
		{at: 84*time.Hour + time.Second, renewalDue: true, description: "into the second half"},
		{at: 7*24*time.Hour - time.Nanosecond, renewalDue: true, description: "about to end"},
		{at: 7 * 24 * time.Hour, expired: true, renewalDue: true, description: "at the deadline"},
	}
	for _, tt := range tests {
		now := start.Add(tt.at)
		if got := s.Expired(now); got != tt.expired {
			t.Errorf("%s: Expired() = %v, want %v", tt.description, got, tt.expired)
		}
		if got := s.DueForRenewal(now); got != tt.renewalDue {
			t.Errorf("%s: DueForRenewal() = %v, want %v", tt.description, got, tt.renewalDue)
		}
	}
}

func TestHashTokenIsStable(t *testing.T) {
	first, again := account.HashToken("a"), account.HashToken("a")
	if first != again {
		t.Error("the same token hashed twice gave two hashes")
	}
	if first == account.HashToken("b") {
		t.Error("two tokens gave the same hash")
	}
}
