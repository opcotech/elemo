package pkg

import "math/rand"

func randomRune(runes []rune) rune {
	return runes[rand.Intn(len(runes))] // #nosec
}

// GenerateRandomStringNumeric generates a random string of length n.
//
// The string is generated using the following characters:
//
//	0123456789
//
// The length of the string is determined by the n parameter.
//
// The string is not guaranteed to be unique, but with a large enough n, the
// probability of a collision is very low.
func GenerateRandomStringNumeric(n int) string {
	runes := []rune("0123456789")
	b := make([]rune, n)

	for i := range b {
		b[i] = randomRune(runes)
	}

	return string(b)
}

// GenerateRandomString generates a random string of length n.
//
// The string is generated using the following characters:
//
//	abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
//
// The length of the string is determined by the n parameter.
//
// The string is not guaranteed to be unique, but with a large enough n, the
// probability of a collision is very low.
func GenerateRandomString(n int) string {
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)

	for i := range b {
		b[i] = randomRune(runes)
	}

	return string(b)
}

// GenerateRandomStringAlpha generates a random string of length n.
//
// The string is generated using the following characters:
//
//	abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
//
// The length of the string is determined by the n parameter.
//
// The string is not guaranteed to be unique, but with a large enough n, the
// probability of a collision is very low.
func GenerateRandomStringAlpha(n int) string {
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)

	for i := range b {
		b[i] = randomRune(runes)
	}

	return string(b)
}

// GenerateEmail returns a random email address for ending with @example.com.
func GenerateEmail(n int) string {
	return GenerateRandomString(n) + "@example.com"
}
