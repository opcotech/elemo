package model

const (
	PermissionKindAll    PermissionKind = iota + 1 // *
	PermissionKindCreate                           // create
	PermissionKindRead                             // read
	PermissionKindWrite                            // write
	PermissionKindDelete                           // delete
)

// PermissionKind represents a permission attached to a relation.
//
//go:generate go tool enumer -type=PermissionKind -text -transform=noop -linecomment -output=permission_kind_gen.go
type PermissionKind uint8
