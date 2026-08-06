package model

import (
	"database/sql/driver"
)

const (
	UserTokenContextConfirm UserTokenContext = iota + 1
	UserTokenContextResetPassword
	UserTokenContextInvite
)

var (
	userTokenContextKeys = map[string]UserTokenContext{
		"confirm":        UserTokenContextConfirm,
		"reset_password": UserTokenContextResetPassword,
		"invite":         UserTokenContextInvite,
	}
	userTokenContextValues = map[UserTokenContext]string{
		UserTokenContextConfirm:       "confirm",
		UserTokenContextResetPassword: "reset_password",
		UserTokenContextInvite:        "invite",
	}
)

// UserTokenContext represents the reason of user token creation.
type UserTokenContext uint8

// String returns the string representation of the LinkStatus.
func (c UserTokenContext) String() string {
	return userTokenContextValues[c]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (c UserTokenContext) MarshalText() (text []byte, err error) {
	if c < UserTokenContextConfirm || c > UserTokenContextInvite {
		return nil, ErrInvalidUserTokenContext
	}
	return []byte(c.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (c *UserTokenContext) UnmarshalText(text []byte) error {
	if v, ok := userTokenContextKeys[string(text)]; ok {
		*c = v
		return nil
	}
	return ErrInvalidUserTokenContext
}

// Scan DB value to LinkStatus.
func (c *UserTokenContext) Scan(value any) error {
	return c.UnmarshalText([]byte(value.(string)))
}

// Value returns the DB compatible value.
func (c UserTokenContext) Value() (driver.Value, error) {
	status, err := c.MarshalText()
	return string(status), err
}
