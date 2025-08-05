package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &PasswordResetTemplateData{
		Subject:          "[Action Required] Reset your password",
		FirstName:        "Test",
		LastName:         "Bob",
		PasswordResetURL: "https://example.com/password/reset",
		SupportEmail:     "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
