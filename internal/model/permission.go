package model

const (
	PermissionKindAll    PermissionKind = iota + 1 // permission to do everything on a resource
	PermissionKindCreate                           // permission to create a resource
	PermissionKindRead                             // permission to read a resource
	PermissionKindWrite                            // permission to write a resource
	PermissionKindDelete                           // permission to delete a resource
)

// PermissionKind represents a permission attached to a relation.
//
//go:generate go tool enumer -type=PermissionKind -trimprefix=PermissionKind -text -transform=snake -output=permission_kind_gen.go
type PermissionKind uint8
