package email

// OrganizationInviteTemplateData represents the data needed to render the
// invitation email template.
type OrganizationInviteTemplateData struct {
	Subject          string `validate:"required,min=3,max=170"`
	FirstName        string `validate:"required,min=3,max=50"`
	LastName         string `validate:"required,min=3,max=50"`
	OrganizationName string `validate:"required,min=3,max=120"`
	InvitationURL    string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Get returns the invitation email template data.
func (d *OrganizationInviteTemplateData) Get() any {
	return d
}
