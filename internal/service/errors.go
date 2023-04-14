package service

import "errors"

var (
	ErrInvalidEmail            = errors.New("invalid email address")             // invalid email address
	ErrInvalidPaginationParams = errors.New("invalid pagination parameters")     // invalid pagination parameters
	ErrNoLicenseService        = errors.New("no license service provided")       // no license service provided
	ErrNoPatchData             = errors.New("no patch data provided")            // no patch data provided
	ErrNoPermission            = errors.New("no permission")                     // no permission
	ErrNoPermissionRepository  = errors.New("no permission repository provided") // no permission repository provided
	ErrNoResources             = errors.New("no resources provided")             // no resources provided
	ErrNoUser                  = errors.New("no user provided")                  // no user provided
	ErrNoUserRepository        = errors.New("no user repository provided")       // no user repository provided
	ErrNoVersionInfo           = errors.New("no version info provided")          // no version info provided
	ErrSystemHealthCheck       = errors.New("system health check failed")        // system health check failed
	ErrUserCreate              = errors.New("failed to create user")             // failed to create user
	ErrUserDelete              = errors.New("failed to delete user")             // failed to delete user
	ErrUserGet                 = errors.New("failed to get user")                // failed to get user
	ErrUserGetAll              = errors.New("failed to get users")               // failed to get users
	ErrUserUpdate              = errors.New("failed to update user")             // failed to update user
	ErrQuotaUsageGet           = errors.New("failed to get usage of quota")      // failed to get usage of quota
	ErrQuotaInvalid            = errors.New("invalid quota")                     // invalid quota
	ErrQuotaExceeded           = errors.New("quota exceeded")                    // quota exceeded
)
