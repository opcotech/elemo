package smtp

import "errors"

var (
	ErrNoSMTPClient   = errors.New("no SMTP client provided") // no SMTP client provided
	ErrAuthFailed     = errors.New("authentication failed")   // authentication failed
	ErrComposeEmail   = errors.New("failed to compose email") // failed to compose email
	ErrNoBytesWritten = errors.New("no bytes written")        // no bytes written
)
