package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	UnusablePassword = "- unusable -" // #nosec password that can't be used for login
)

// HashPassword creates a hash from the password.
func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// IsPasswordMatching validates that the raw password matches the hashed one.
func IsPasswordMatching(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
