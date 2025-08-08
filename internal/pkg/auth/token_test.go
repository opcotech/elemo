package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestIsTokenMatching(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		token    string
		expected bool
	}{
		{
			name:     "test token hash matches",
			hash:     "$2a$10$2/R2NpjFJRbFFMKNBkBzoORxEfwiBwnWEQ5yDdU6H1rY/quJn2lUO",
			token:    "Y29uZmlybTtwMTdqSDAza2RPNWR3MHNLcTFiYjNDSWVjUlhzUXFuSWx6Wkw7eyJkYXRhIjoidGVzdCJ9",
			expected: true,
		},
		{
			name:     "test token hash not matches",
			hash:     "$2a$10$6cO/7Nn9uxkgZbS.6cVVA.vdrcMyjycAE1o4ysT2/FZWt/WVxtVhq",
			token:    "NOT-MATCHING",
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := GenerateToken(model.UserTokenContextConfirm.String(), map[string]any{
				"data": "test",
			})
			require.NoError(t, err)

			isMatching := IsTokenMatching(tt.hash, tt.token)
			require.Equal(t, tt.expected, isMatching)
		})
	}
}
