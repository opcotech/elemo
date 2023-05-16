package email

import (
	"errors"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// OrganizationInviteTemplateData represents the data needed to render the
// invitation email template.
type OrganizationInviteTemplateData struct {
	Username         string `validate:"required,min=3,max=50"`
	FirstName        string `validate:"omitempty,max=50"`
	OrganizationName string `validate:"required,min=3,max=50"`
	InvitationURL    string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Validate validates the invitation email template data.
func (d *OrganizationInviteTemplateData) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidOrganizationInviteTemplateData, err)
	}
	return nil
}

// Get returns the invitation email template data.
func (d *OrganizationInviteTemplateData) Get() interface{} {
	return d
}

// NewOrganizationInviteTemplateData returns a new OrganizationInviteTemplateData struct.
func NewOrganizationInviteTemplateData(username, firstName, organizationName, invitationURL, supportEmail string) (*OrganizationInviteTemplateData, error) {
	d := &OrganizationInviteTemplateData{
		Username:         username,
		FirstName:        firstName,
		OrganizationName: organizationName,
		InvitationURL:    invitationURL,
		SupportEmail:     supportEmail,
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return d, nil
}
