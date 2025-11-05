package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationInviteTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &OrganizationInviteTemplateData{
		Subject:          "You have been invited to join an organization",
		OrganizationName: "ACME Inc.",
		InvitationURL:    "https://example.com/org/1/invite",
		SupportEmail:     "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
