package smtp

import "errors"

var (
	ErrComposeEmail  = errors.New("failed to compose email") // failed to compose email
	ErrInvalidClient = errors.New("invalid SMTP client")     // invalid SMTP client
	ErrInvalidEmail  = errors.New("invalid email")           // invalid email
	ErrSendEmail     = errors.New("failed to send email")    // failed to send email
)
