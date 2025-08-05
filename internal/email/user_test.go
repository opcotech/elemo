package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserWelcomeTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &UserWelcomeTemplateData{
		Subject:      "Welcome to Elemo!",
		FirstName:    "Test",
		LastName:     "Bob",
		LoginURL:     "https://example.com",
		SupportEmail: "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
