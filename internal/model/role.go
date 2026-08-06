package model

const (
	SystemRoleOwner   SystemRole = iota + 1 // the owner them of the instance
	SystemRoleAdmin                         // the administrator team of the instance
	SystemRoleSupport                       // the support team of the instance
)

var (
	systemRoleValues = map[SystemRole]string{
		SystemRoleOwner:   "Owner",
		SystemRoleAdmin:   "Admin",
		SystemRoleSupport: "Support",
	}
	systemRoleKeys = map[string]SystemRole{
		"Owner":   SystemRoleOwner,
		"Admin":   SystemRoleAdmin,
		"Support": SystemRoleSupport,
	}
)

// SystemRole is a special role that is created by the system.
type SystemRole uint8

// String returns the string representation of the SystemRole.
func (r SystemRole) String() string {
	return systemRoleValues[r]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (r SystemRole) MarshalText() (text []byte, err error) {
	if r < 1 || r > 3 {
		return nil, ErrInvalidSystemRole
	}
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *SystemRole) UnmarshalText(text []byte) error {
	if v, ok := systemRoleKeys[string(text)]; ok {
		*r = v
		return nil
	}
	return ErrInvalidSystemRole
}
