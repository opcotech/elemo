package email

import (
	"github.com/opcotech/elemo/internal/license"
)

// LicenseExpiryTemplateData represents the data needed to render the password
// reset email template.
type LicenseExpiryTemplateData struct {
	Username     string           `validate:"required,min=3,max=50"`
	FirstName    string           `validate:"omitempty,max=50"`
	License      *license.License `validate:"required"`
	ServerURL    string           `validate:"required,url"`
	RenewEmail   string           `validate:"required,email"`
	SupportEmail string           `validate:"required,email"`
}

// Get returns the license expiration email template data.
func (d *LicenseExpiryTemplateData) Get() interface{} {
	return d
}
