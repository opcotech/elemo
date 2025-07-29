package pkg

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomStringNumeric(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "generate numeric string of length 1",
			length: 1,
		},
		{
			name:   "generate numeric string of length 5",
			length: 5,
		},
		{
			name:   "generate numeric string of length 10",
			length: 10,
		},
		{
			name:   "generate numeric string of length 0",
			length: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateRandomStringNumeric(tt.length)
			require.Len(t, result, tt.length)

			// Check that the string contains only numeric characters
			matched, err := regexp.MatchString(`^[0-9]*$`, result)
			require.NoError(t, err)
			assert.True(t, matched, "String should contain only numeric characters")
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "generate random string of length 1",
			length: 1,
		},
		{
			name:   "generate random string of length 5",
			length: 5,
		},
		{
			name:   "generate random string of length 10",
			length: 10,
		},
		{
			name:   "generate random string of length 0",
			length: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateRandomString(tt.length)
			require.Len(t, result, tt.length)

			// Check that the string contains only alphanumeric characters
			matched, err := regexp.MatchString(`^[a-zA-Z0-9]*$`, result)
			require.NoError(t, err)
			assert.True(t, matched, "String should contain only alphanumeric characters")
		})
	}
}

func TestGenerateRandomStringAlpha(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected string
	}{
		{
			name:   "generate alpha string of length 1",
			length: 1,
		},
		{
			name:   "generate alpha string of length 5",
			length: 5,
		},
		{
			name:   "generate alpha string of length 10",
			length: 10,
		},
		{
			name:   "generate alpha string of length 0",
			length: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateRandomStringAlpha(tt.length)
			require.Len(t, result, tt.length)

			// Check that the string contains only alphabetic characters
			matched, err := regexp.MatchString(`^[a-zA-Z]*$`, result)
			require.NoError(t, err)
			assert.True(t, matched, "String should contain only alphabetic characters")
		})
	}
}

func TestGenerateEmail(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "generate email with length 1",
			length: 1,
		},
		{
			name:   "generate email with length 5",
			length: 5,
		},
		{
			name:   "generate email with length 10",
			length: 10,
		},
		{
			name:   "generate email with length 0",
			length: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateEmail(tt.length)

			// Check that the email ends with @example.com
			assert.True(t, strings.HasSuffix(result, "@example.com"), "Email should end with @example.com")

			// Check that the local part has the correct length
			localPart := strings.TrimSuffix(result, "@example.com")
			require.Len(t, localPart, tt.length)

			// Check that the local part contains only alphanumeric characters
			if tt.length > 0 {
				matched, err := regexp.MatchString(`^[a-zA-Z0-9]*$`, localPart)
				require.NoError(t, err)
				assert.True(t, matched, "Local part should contain only alphanumeric characters")
			}
		})
	}
}

func TestRandomness(t *testing.T) {
	t.Run("numeric strings should be different", func(t *testing.T) {
		t.Parallel()
		// Generate multiple strings and check they're different
		results := make(map[string]bool)
		for range 100 {
			result := GenerateRandomStringNumeric(10)
			if results[result] {
				t.Fatalf("Duplicate numeric string generated: %s", result)
			}
			results[result] = true
		}
	})

	t.Run("random strings should be different", func(t *testing.T) {
		t.Parallel()
		// Generate multiple strings and check they're different
		results := make(map[string]bool)
		for range 100 {
			result := GenerateRandomString(10)
			if results[result] {
				t.Fatalf("Duplicate random string generated: %s", result)
			}
			results[result] = true
		}
	})

	t.Run("alpha strings should be different", func(t *testing.T) {
		t.Parallel()
		// Generate multiple strings and check they're different
		results := make(map[string]bool)
		for range 100 {
			result := GenerateRandomStringAlpha(10)
			if results[result] {
				t.Fatalf("Duplicate alpha string generated: %s", result)
			}
			results[result] = true
		}
	})

	t.Run("emails should be different", func(t *testing.T) {
		t.Parallel()
		// Generate multiple emails and check they're different
		results := make(map[string]bool)
		for range 100 {
			result := GenerateEmail(10)
			if results[result] {
				t.Fatalf("Duplicate email generated: %s", result)
			}
			results[result] = true
		}
	})
}

func TestCharacterSets(t *testing.T) {
	t.Run("numeric strings should only contain digits", func(t *testing.T) {
		t.Parallel()
		result := GenerateRandomStringNumeric(100)
		for _, char := range result {
			assert.True(t, char >= '0' && char <= '9', "Character %c should be a digit", char)
		}
	})

	t.Run("random strings should only contain alphanumeric characters", func(t *testing.T) {
		t.Parallel()
		result := GenerateRandomString(100)
		for _, char := range result {
			assert.True(t,
				(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9'),
				"Character %c should be alphanumeric", char)
		}
	})

	t.Run("alpha strings should only contain alphabetic characters", func(t *testing.T) {
		t.Parallel()
		result := GenerateRandomStringAlpha(100)
		for _, char := range result {
			assert.True(t,
				(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z'),
				"Character %c should be alphabetic", char)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("negative length should panic", func(t *testing.T) {
		t.Parallel()
		// The current implementation panics with negative lengths
		// This test documents the current behavior
		assert.Panics(t, func() {
			GenerateRandomStringNumeric(-1)
		})
	})

	t.Run("very large length should work", func(t *testing.T) {
		t.Parallel()
		result := GenerateRandomString(1000)
		assert.Len(t, result, 1000)

		// Check that it contains only alphanumeric characters
		matched, err := regexp.MatchString(`^[a-zA-Z0-9]*$`, result)
		require.NoError(t, err)
		assert.True(t, matched)
	})
}

func TestConsistency(t *testing.T) {
	t.Run("same length should produce consistent results", func(t *testing.T) {
		t.Parallel()
		length := 10

		// Generate multiple strings of the same length
		results := make([]string, 10)
		for i := 0; i < 10; i++ {
			results[i] = GenerateRandomString(length)
		}

		// All should have the same length
		for _, result := range results {
			assert.Len(t, result, length)
		}
	})

	t.Run("email consistency", func(t *testing.T) {
		t.Parallel()
		length := 5

		// Generate multiple emails of the same length
		results := make([]string, 10)
		for i := 0; i < 10; i++ {
			results[i] = GenerateEmail(length)
		}

		// All should have the same total length
		expectedLength := length + len("@example.com")
		for _, result := range results {
			assert.Len(t, result, expectedLength)
			assert.True(t, strings.HasSuffix(result, "@example.com"))
		}
	})
}
