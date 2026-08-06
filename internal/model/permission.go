package model

const (
	PermissionKindAll    PermissionKind = iota + 1 // permission to do everything on a resource
	PermissionKindCreate                           // permission to create a resource
	PermissionKindRead                             // permission to read a resource
	PermissionKindWrite                            // permission to write a resource
	PermissionKindDelete                           // permission to delete a resource
)

var (
	permissionKindKeys = map[string]PermissionKind{
		"*":      PermissionKindAll,
		"create": PermissionKindCreate,
		"read":   PermissionKindRead,
		"write":  PermissionKindWrite,
		"delete": PermissionKindDelete,
	}
	permissionKindValues = map[PermissionKind]string{
		PermissionKindAll:    "*",
		PermissionKindCreate: "create",
		PermissionKindRead:   "read",
		PermissionKindWrite:  "write",
		PermissionKindDelete: "delete",
	}
)

// PermissionKind represents a permission attached to a relation.
type PermissionKind uint8

// String returns the string representation of the permission.
func (p PermissionKind) String() string {
	return permissionKindValues[p]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p PermissionKind) MarshalText() (text []byte, err error) {
	if p < 1 || p > 5 {
		return nil, ErrInvalidPermissionKind
	}
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *PermissionKind) UnmarshalText(text []byte) error {
	if v, ok := permissionKindKeys[string(text)]; ok {
		*p = v
		return nil
	}
	return ErrInvalidPermissionKind
}
