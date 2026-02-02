package auth

import (
	"gophermart/internal/gophermart/application/port"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

// BCryptHasher hashes and verifies passwords with bcrypt.
type BCryptHasher struct{}

// NewBCryptHasher returns a new bcrypt hasher.
func NewBCryptHasher() *BCryptHasher {
	return &BCryptHasher{}
}

// Hash hashes the plain password.
func (h *BCryptHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare returns true if plain matches hash.
func (h *BCryptHasher) Compare(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

var _ port.PasswordHasher = (*BCryptHasher)(nil)
