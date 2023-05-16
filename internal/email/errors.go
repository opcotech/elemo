package email

import "errors"

var (
	ErrInvalidLicenseExpiryTemplateData      = errors.New("invalid license expiration template data")  // invalid license expiration template data
	ErrInvalidOrganizationInviteTemplateData = errors.New("invalid organization invite template data") // invalid organization invite template data
	ErrInvalidPasswordResetTemplateData      = errors.New("invalid password reset template data")      // invalid password reset template data
	ErrInvalidUserWelcomeTemplateData        = errors.New("invalid welcome template data")             // invalid welcome template data
	ErrTemplateExecute                       = errors.New("failed to execute email template")          // failed to execute email template
	ErrTemplateInvalid                       = errors.New("invalid template")                          // invalid template
	ErrTemplateParse                         = errors.New("failed to parse email template")            // failed to parse email template
)
