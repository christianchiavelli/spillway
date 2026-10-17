// Package argon2id hashes passwords with Argon2id, in the PHC string format
// every library reads: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>.
package argon2id

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params are the costs of one hash.
type Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  int
	KeyLength   uint32
}

// DefaultParams are OWASP's first recommendation for Argon2id, checked in
// October 2026: 19 MiB of memory, two passes, one lane.
var DefaultParams = Params{MemoryKiB: 19 * 1024, Iterations: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32}

// ErrMalformed is returned for a stored hash this package cannot read.
var ErrMalformed = errors.New("argon2id: malformed hash")

// Hasher hashes and checks passwords. Each hash holds MemoryKiB of memory for
// as long as it runs, so a burst of sign-ins could take the process down;
// slots caps how many run at once, and the rest wait their turn or give up
// with their request.
type Hasher struct {
	params Params
	slots  chan struct{}
}

// New returns a Hasher that runs at most concurrency hashes at a time.
func New(params Params, concurrency int) *Hasher {
	return &Hasher{params: params, slots: make(chan struct{}, max(concurrency, 1))}
}

// Hash returns the encoded hash of password with a new random salt.
func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	rand.Read(salt) // never fails, since Go 1.24
	key, err := h.derive(ctx, password, salt, h.params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.MemoryKiB, h.params.Iterations, h.params.Parallelism,
		b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// Verify reports whether password matches encoded. It reads the costs from
// encoded, so hashes made before a change of Params still verify.
func (h *Hasher) Verify(ctx context.Context, password, encoded string) (bool, error) {
	params, salt, want, err := decode(encoded)
	if err != nil {
		return false, err
	}
	got, err := h.derive(ctx, password, salt, params)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func (h *Hasher) derive(ctx context.Context, password string, salt []byte, p Params) ([]byte, error) {
	select {
	case h.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("argon2id: waiting for a free slot: %w", context.Cause(ctx))
	}
	defer func() { <-h.slots }()
	return argon2.IDKey([]byte(password), salt, p.Iterations, p.MemoryKiB, p.Parallelism, p.KeyLength), nil
}

var b64 = base64.RawStdEncoding

func decode(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, ErrMalformed
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Params{}, nil, nil, ErrMalformed
	}
	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.MemoryKiB, &p.Iterations, &p.Parallelism); err != nil {
		return Params{}, nil, nil, ErrMalformed
	}
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, ErrMalformed
	}
	key, err := b64.DecodeString(parts[5])
	if err != nil || len(key) == 0 || len(key) > 1024 {
		return Params{}, nil, nil, ErrMalformed
	}
	p.SaltLength, p.KeyLength = len(salt), uint32(len(key)) //nolint:gosec // at most 1024, checked above
	return p, salt, key, nil
}
