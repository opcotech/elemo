package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "test password hashed",
			password: "secret",
		},
		{
			name:     "test empty password",
			password: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, HashPassword(tt.password))
		})
	}
}

func TestIsPasswordMatching(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		password string
		expected bool
	}{
		{
			name:     "test password hash matches",
			hash:     "$2a$10$x7EeYY7WdIwpmBemnfbbGeAR99tLEJ9Ig7LSF.IqFzbe06iR4X6Gq",
			password: "secret",
			expected: true,
		},
		{
			name:     "test password hash matches for the same password",
			hash:     "$2a$10$dl9wdqFAflo0GLp.hJ5EXO.vHcLd8eNFt8Z/KSQH0bJvvIDtPw69y",
			password: "secret",
			expected: true,
		},
		{
			name:     "test password hash not matches",
			hash:     "$2a$10$dl9wdqFAflo0GLp.hJ5EXO.vHcLd8eNFt8Z/KSQH0bJvvIDtPw69W",
			password: "secret",
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isMatching := IsPasswordMatching(tt.hash, tt.password)
			require.Equal(t, tt.expected, isMatching)
		})
	}
}
