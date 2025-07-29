package testutil

import "github.com/opcotech/elemo/internal/pkg"

// GenerateEmail returns a random email address for ending with @example.com.
func GenerateEmail(n int) string {
	return pkg.GenerateRandomString(n) + "@example.com"
}
