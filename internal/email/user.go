package email

// UserConfirmationTemplateData represents the data needed to render the
// confirmation email template.
type UserConfirmationTemplateData struct {
	Subject      string `validate:"required,min=3,max=50"`
	FirstName    string `validate:"required,min=1,max=50"`
	LastName     string `validate:"required,min=1,max=50"`
	VerifyURL    string `validate:"required,url"`
	SupportEmail string `validate:"required,email"`
}

// Get returns the welcome email template data.
func (d *UserConfirmationTemplateData) Get() any {
	return d
}

// UserWelcomeTemplateData represents the data needed to render the welcome
// email template.
type UserWelcomeTemplateData struct {
	Subject      string `validate:"required,min=3,max=50"`
	FirstName    string `validate:"required,min=1,max=50"`
	LastName     string `validate:"required,min=1,max=50"`
	LoginURL     string `validate:"required,url"`
	SupportEmail string `validate:"required,email"`
}

// Get returns the welcome email template data.
func (d *UserWelcomeTemplateData) Get() any {
	return d
}
