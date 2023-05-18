package email

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/license"
)

func TestLicenseExpiryTemplateData_Get(t *testing.T) {
	t.Parallel()

	l := license.License{
		ID:           xid.New(),
		Email:        "info@example.com",
		Organization: "ACME Inc.",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	data := &LicenseExpiryTemplateData{
		Subject:             "License is about to expire",
		Username:            "test-user",
		FirstName:           "Test",
		LicenseID:           l.ID.String(),
		LicenseEmail:        l.Email,
		LicenseOrganization: l.Organization,
		LicenseExpiresAt:    l.ExpiresAt.Format(time.RFC850),
		ServerURL:           "",
		RenewEmail:          "renew@example.com",
		SupportEmail:        "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
