package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &PasswordResetTemplateData{
		Subject:          "Reset your password",
		Username:         "test-user",
		FirstName:        "Test",
		PasswordResetURL: "https://example.com/password/reset",
		SupportEmail:     "info@example.com",
	}

	assert.Equal(t, data, data.RenderData())
}
