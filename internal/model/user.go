package model

const (
	UserStatusActive   UserStatus = iota + 1 // active
	UserStatusPending                        // pending
	UserStatusInactive                       // inactive
	UserStatusDeleted                        // deleted
)

// UserStatus represents the status of the User in the system.
//
//go:generate go tool enumer -type=UserStatus -text -transform=noop -linecomment -output=user_status_gen.go
type UserStatus uint8
