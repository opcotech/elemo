package model

const (
	SystemRoleOwner   SystemRole = iota + 1 // the owner them of the instance
	SystemRoleAdmin                         // the administrator team of the instance
	SystemRoleSupport                       // the support team of the instance
)

// SystemRole is a special role that is created by the system.
//
//go:generate go tool enumer -type=SystemRole -trimprefix=SystemRole -text -transform=noop -output=role_system_role_gen.go
type SystemRole uint8
