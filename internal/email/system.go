package email

// LicenseExpiryTemplateData represents the data needed to render the password
// reset email template.
type LicenseExpiryTemplateData struct {
	Subject             string `validate:"required,min=3,max=50"`
	LicenseID           string `validate:"required"`
	LicenseEmail        string `validate:"required,email"`
	LicenseOrganization string `validate:"required"`
	LicenseExpiresAt    string `validate:"required"`
	ServerURL           string `validate:"required,url"`
	RenewEmail          string `validate:"required,email"`
	SupportEmail        string `validate:"required,email"`
}

// Get returns the license expiration email template data.
func (d *LicenseExpiryTemplateData) Get() interface{} {
	return d
}
