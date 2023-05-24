package http

import "errors"

var (
	ErrAuthCredentials       = errors.New("invalid credentials")              // invalid credentials
	ErrAuthNoPermission      = errors.New("no permission")                    // no permission
	ErrInvalidSwagger        = errors.New("invalid swagger provided")         // invalid swagger provided
	ErrNoAuthProvider        = errors.New("no auth provider provided")        // no auth provider provided
	ErrNoLicenseService      = errors.New("no license service provided")      // no license service provided
	ErrNoLogger              = errors.New("no logger provided")               // no logger provided
	ErrNoOrganizationService = errors.New("no organization service provided") // no organization service provided
	ErrNoSystemService       = errors.New("no system service provided")       // no system service provided
	ErrNoTodoService         = errors.New("no todo service provided")         // no todo service provided
	ErrNoTracer              = errors.New("no tracer provided")               // no tracer provided
	ErrNoUserService         = errors.New("no user service provided")         // no user service provided
)
