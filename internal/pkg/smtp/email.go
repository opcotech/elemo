package smtp

import (
	"errors"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// Template represents an email template.
type Template interface {
	// Body returns the email body.
	Body() (string, error)
	// MimeType returns the email mime type.
	MimeType() string
	// Validate validates the email template.
	Validate() error
}

// Email is a struct that represents a plain text email message.
type Email struct {
	Subject  string   `validate:"required"`
	From     string   `validate:"required,email"`
	To       []string `validate:"required,min=1,dive,email"`
	Template Template `validate:"required"`
}

// MimeHeader returns the email mime header.
func (e *Email) MimeHeader() string {
	return "MIME-version: 1.0;\nContent-Type: " + e.Template.MimeType() + "; charset=\"UTF-8\";\n\n"
}

// Validate validates the email.
func (e *Email) Validate() error {
	if err := validate.Struct(e); err != nil {
		return errors.Join(ErrInvalidEmail, err)
	}
	return e.Template.Validate()
}

// NewEmail returns a new instance of an email.
func NewEmail(subject, from string, to []string, template Template) (*Email, error) {
	e := &Email{
		Subject:  subject,
		From:     from,
		To:       to,
		Template: template,
	}

	if err := e.Validate(); err != nil {
		return nil, err
	}

	return e, nil
}
