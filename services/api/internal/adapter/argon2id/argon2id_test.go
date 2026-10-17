package argon2id_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/argon2id"
)

// cheap keeps the suite fast; the format and the checks are the same.
var cheap = argon2id.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func TestHashAndVerify(t *testing.T) {
	h := argon2id.New(cheap, 2)

	encoded, err := h.Hash(t.Context(), "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Errorf("encoded = %q, want the PHC format", encoded)
	}

	for password, want := range map[string]bool{
		"correct horse battery staple":  true,
		"correct horse battery staple ": false,
		"":                              false,
	} {
		got, err := h.Verify(t.Context(), password, encoded)
		if err != nil || got != want {
			t.Errorf("Verify(%q) = %v, %v; want %v", password, got, err, want)
		}
	}
}

func TestHashSaltsEachPassword(t *testing.T) {
	h := argon2id.New(cheap, 1)
	a, _ := h.Hash(t.Context(), "same password here")
	b, _ := h.Hash(t.Context(), "same password here")
	if a == b {
		t.Error("two hashes of one password are equal")
	}
}

func TestVerifyReadsTheCostsFromTheHash(t *testing.T) {
	old, err := argon2id.New(cheap, 1).Hash(t.Context(), "made before the change")
	if err != nil {
		t.Fatal(err)
	}
	stronger := cheap
	stronger.Iterations = 3

	ok, err := argon2id.New(stronger, 1).Verify(t.Context(), "made before the change", old)
	if err != nil || !ok {
		t.Errorf("Verify() with new costs = %v, %v; want true", ok, err)
	}
}

func TestVerifyRejectsMalformedHashes(t *testing.T) {
	h := argon2id.New(cheap, 1)
	for _, encoded := range []string{
		"",
		"plain",
		"$argon2i$v=19$m=64,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=16$m=64,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=x,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=64,t=1,p=1$not base64!$aGFzaA",
		"$argon2id$v=19$m=64,t=1,p=1$c2FsdA$",
	} {
		if _, err := h.Verify(t.Context(), "anything", encoded); !errors.Is(err, argon2id.ErrMalformed) {
			t.Errorf("Verify(%q) error = %v, want ErrMalformed", encoded, err)
		}
	}
}
