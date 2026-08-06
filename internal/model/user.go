package model

const (
	UserStatusActive   UserStatus = iota + 1 // the user is active
	UserStatusPending                        // the user is invited but not yet active
	UserStatusInactive                       // the user is inactive
	UserStatusDeleted                        // the user is deleted
)

var (
	userStatusKeys = map[string]UserStatus{
		"active":   UserStatusActive,
		"pending":  UserStatusPending,
		"inactive": UserStatusInactive,
		"deleted":  UserStatusDeleted,
	}
	userStatusValues = map[UserStatus]string{
		UserStatusActive:   "active",
		UserStatusPending:  "pending",
		UserStatusInactive: "inactive",
		UserStatusDeleted:  "deleted",
	}
)

// UserStatus represents the status of the User in the system.
type UserStatus uint8

// String returns the string representation of the UserStatus.
func (s UserStatus) String() string {
	return userStatusValues[s]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s UserStatus) MarshalText() (text []byte, err error) {
	if s < 1 || s > 4 {
		return nil, ErrInvalidUserStatus
	}
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *UserStatus) UnmarshalText(text []byte) error {
	if v, ok := userStatusKeys[string(text)]; ok {
		*s = v
		return nil
	}
	return ErrInvalidUserStatus
}
