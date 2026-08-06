package model

const (
	UserStatusActive   UserStatus = iota + 1 // the user is active
	UserStatusPending                        // the user is invited but not yet active
	UserStatusInactive                       // the user is inactive
	UserStatusDeleted                        // the user is deleted
)

// UserStatus represents the status of the User in the system.
//
//go:generate go tool enumer -type=UserStatus -trimprefix=UserStatus -text -transform=snake -output=user_status_gen.go
type UserStatus uint8
