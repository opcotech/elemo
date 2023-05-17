package email

// UserWelcomeTemplateData represents the data needed to render the welcome
// email template.
type UserWelcomeTemplateData struct {
	Subject      string `validate:"required,min=3,max=50"`
	Username     string `validate:"required,min=3,max=50"`
	FirstName    string `validate:"omitempty,max=50"`
	LoginURL     string `validate:"required,url"`
	SupportEmail string `validate:"required,email"`
}

// Get returns the welcome email template data.
func (d *UserWelcomeTemplateData) Get() interface{} {
	return d
}
