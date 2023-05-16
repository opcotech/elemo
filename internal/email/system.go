package email

import (
	"errors"

	"github.com/rs/xid"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

// LicenseExpiryTemplateData represents the data needed to render the password
// reset email template.
type LicenseExpiryTemplateData struct {
	Username     string           `validate:"required,min=3,max=50"`
	FirstName    string           `validate:"omitempty,max=50"`
	License      *license.License `validate:"required"`
	RenewEmail   string           `validate:"required,email"`
	SupportEmail string           `validate:"required,email"`
}

// Validate validates the license expiration email template data.
func (d *LicenseExpiryTemplateData) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidLicenseExpiryTemplateData, err)
	}
	if d.License.ID == xid.NilID() {
		return errors.Join(ErrInvalidLicenseExpiryTemplateData, license.ErrLicenseInvalid)
	}
	return nil
}

// Get returns the license expiration email template data.
func (d *LicenseExpiryTemplateData) Get() interface{} {
	return d
}

// NewLicenseExpiryTemplateData returns a new LicenseExpiryTemplateData struct.
func NewLicenseExpiryTemplateData(username, firstName, renewEmail, supportEmail string, license *license.License) (*LicenseExpiryTemplateData, error) {
	d := &LicenseExpiryTemplateData{
		Username:     username,
		FirstName:    firstName,
		License:      license,
		RenewEmail:   renewEmail,
		SupportEmail: supportEmail,
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return d, nil
}
