package email

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseExpiryTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &LicenseExpiryTemplateData{
		Subject:        "License expiration reminder",
		Customer:       "ACME Inc.",
		LicenseID:      "license-id",
		LicenseState:   "valid",
		LicenseExpires: "15 Sep 2026 09:00 UTC",
		GraceEnds:      "15 Oct 2026 09:00 UTC",
		SeatsLicensed:  10,
		SettingsURL:    "https://example.com/settings",
		SupportEmail:   "support@example.com",
	}

	assert.Equal(t, data, data.Get())
}

func TestLicenseExpiryTemplateRenders(t *testing.T) {
	t.Parallel()

	tmpl, err := NewTemplate(
		filepath.Join("..", "..", "templates", "email", "license-expiry-reminder.html"),
		&LicenseExpiryTemplateData{
			Subject:        "License expiration reminder",
			Customer:       "ACME Inc.",
			LicenseID:      "license-id",
			LicenseState:   "grace",
			LicenseExpires: "15 Sep 2026 09:00 UTC",
			GraceEnds:      "15 Oct 2026 09:00 UTC",
			SeatsLicensed:  10,
			SettingsURL:    "https://example.com/settings",
			SupportEmail:   "support@example.com",
		},
	)
	require.NoError(t, err)

	body, err := tmpl.Render()
	require.NoError(t, err)
	for _, want := range []string{
		"ACME Inc.",
		"license-id",
		"grace",
		"15 Sep 2026 09:00 UTC",
		"15 Oct 2026 09:00 UTC",
		"https://example.com/settings",
	} {
		assert.True(t, strings.Contains(body, want), "rendered email should contain %q", want)
	}
}
