package email

// LicenseExpiryTemplateData represents the data needed to render an AirGap
// license expiration reminder.
type LicenseExpiryTemplateData struct {
	Subject        string `validate:"required,min=3,max=170"`
	Customer       string `validate:"required,max=256"`
	LicenseID      string `validate:"required"`
	LicenseState   string `validate:"required"`
	LicenseExpires string `validate:"required"`
	GraceEnds      string `validate:"required"`
	SeatsLicensed  int    `validate:"gte=0"`
	SettingsURL    string `validate:"required,url"`
	SupportEmail   string `validate:"required,email"`
}

// Get returns the license expiration email template data.
func (d *LicenseExpiryTemplateData) Get() any {
	return d
}
