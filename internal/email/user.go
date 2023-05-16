package email

import (
	"errors"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// UserWelcomeTemplateData represents the data needed to render the welcome
// email template.
type UserWelcomeTemplateData struct {
	Username     string `validate:"required,min=3,max=50"`
	FirstName    string `validate:"omitempty,max=50"`
	LoginURL     string `validate:"required,url"`
	SupportEmail string `validate:"required,email"`
}

// Validate validates the welcome email template data.
func (d *UserWelcomeTemplateData) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidUserWelcomeTemplateData, err)
	}
	return nil
}

// Get returns the welcome email template data.
func (d *UserWelcomeTemplateData) Get() interface{} {
	return d
}

// NewUserWelcomeTemplateData returns a new UserWelcomeTemplateData struct.
func NewUserWelcomeTemplateData(username, firstName, serverURL, supportEmail string) (*UserWelcomeTemplateData, error) {
	d := &UserWelcomeTemplateData{
		Username:     username,
		FirstName:    firstName,
		LoginURL:     serverURL,
		SupportEmail: supportEmail,
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return d, nil
}
