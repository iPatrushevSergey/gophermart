package port

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(plain string) (hash string, err error)
	Compare(plain, hash string) bool
}
