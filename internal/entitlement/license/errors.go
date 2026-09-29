package license

import "errors"

var (
	ErrLicenseMissing          = errors.New("license file is missing")
	ErrLicenseTooLarge         = errors.New("license file exceeds the maximum allowed size")
	ErrLicenseMalformed        = errors.New("license is malformed")
	ErrLicenseUnknownField     = errors.New("license contains unknown fields")
	ErrLicenseFieldOrder       = errors.New("license fields are not in the required order")
	ErrLicenseUnknownKey       = errors.New("license key id is not trusted")
	ErrLicenseInvalidSignature = errors.New("license signature is invalid")
	ErrLicenseInvalidPayload   = errors.New("license payload is invalid")
	ErrNoTrust                 = errors.New("no trust store provided")
)
