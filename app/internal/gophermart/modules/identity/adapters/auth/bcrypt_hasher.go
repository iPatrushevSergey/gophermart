package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// BCryptHasher hashes and verifies passwords with bcrypt.
type BCryptHasher struct {
	cost int
}

// NewBCryptHasher returns a new bcrypt hasher with the given cost (4-31).
func NewBCryptHasher(cost int) *BCryptHasher {
	return &BCryptHasher{cost: cost}
}

// Hash hashes the plain password.
func (h *BCryptHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare returns true if plain matches hash.
func (h *BCryptHasher) Compare(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
