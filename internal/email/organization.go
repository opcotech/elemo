package email

// OrganizationInviteTemplateData represents the data needed to render the
// invitation email template.
type OrganizationInviteTemplateData struct {
	Subject          string `validate:"required,min=3,max=170"`
	OrganizationName string `validate:"required,min=3,max=120"`
	InvitationURL    string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Get returns the invitation email template data.
func (d *OrganizationInviteTemplateData) Get() any {
	return d
}
