package email

// PasswordResetTemplateData represents the data needed to render the password
// reset email template.
type PasswordResetTemplateData struct {
	Subject          string `validate:"required,min=3,max=50"`
	Username         string `validate:"required,min=3,max=50"`
	FirstName        string `validate:"omitempty,max=50"`
	PasswordResetURL string `validate:"required,url"`
	SupportEmail     string `validate:"required,email"`
}

// Get returns the password reset email template data.
func (d *PasswordResetTemplateData) Get() interface{} {
	return d
}
