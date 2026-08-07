package model

const (
	SystemRoleOwner   SystemRole = iota + 1 // Owner
	SystemRoleAdmin                         // Admin
	SystemRoleSupport                       // Support
)

// SystemRole is a special role that is created by the system.
//
//go:generate go tool enumer -type=SystemRole -text -transform=noop -linecomment -output=role_system_role_gen.go
type SystemRole uint8
