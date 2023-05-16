package email

import (
	"errors"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// PasswordResetTemplateData represents the data needed to render the password
// reset email template.
type PasswordResetTemplateData struct {
	Username         string `validate:"required,min=3,max=50"`
	FirstName        string `validate:"omitempty,max=50"`
	PasswordResetURL string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Validate validates the password reset email template data.
func (d *PasswordResetTemplateData) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidPasswordResetTemplateData, err)
	}
	return nil
}

// Get returns the password reset email template data.
func (d *PasswordResetTemplateData) Get() interface{} {
	return d
}

// NewPasswordResetTemplateData returns a new PasswordResetTemplateData struct.
func NewPasswordResetTemplateData(username, firstName, passwordResetURL, supportEmail string) (*PasswordResetTemplateData, error) {
	d := &PasswordResetTemplateData{
		Username:         username,
		FirstName:        firstName,
		PasswordResetURL: passwordResetURL,
		SupportEmail:     supportEmail,
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return d, nil
}
