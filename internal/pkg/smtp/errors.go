package smtp

import "errors"

var (
	ErrNoSMTPClient = errors.New("no SMTP client provided") // no SMTP client provided
	ErrSendEmail    = errors.New("failed to compose email") // failed to compose email
)
