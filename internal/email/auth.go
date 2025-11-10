package email

// PasswordResetTemplateData represents the data needed to render the password
// reset email template.
type PasswordResetTemplateData struct {
	Subject          string `validate:"required,min=3,max=50"`
	FirstName        string `validate:"required,min=1,max=50"`
	LastName         string `validate:"required,min=1,max=50"`
	PasswordResetURL string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Get returns the password reset email template data.
func (d *PasswordResetTemplateData) Get() any {
	return d
}
