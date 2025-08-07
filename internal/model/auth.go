package model

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	UserTokenContextConfirm UserTokenContext = iota + 1
	UserTokenContextResetPassword
)

var (
	userTokenContextKeys = map[string]UserTokenContext{
		"confirm":        UserTokenContextConfirm,
		"reset_password": UserTokenContextResetPassword,
	}
	userTokenContextValues = map[UserTokenContext]string{
		UserTokenContextConfirm:       "confirm",
		UserTokenContextResetPassword: "reset_password",
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
	if c < UserTokenContextConfirm || c > UserTokenContextResetPassword {
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

// UserToken represents a token that was created for the user.
type UserToken struct {
	ID        ID               `json:"id" validate:"required"`
	UserID    ID               `json:"user_id" validate:"required"`
	SentTo    string           `json:"sent_to" validate:"required,email"`
	Token     string           `json:"token" validate:"required,min=60,max=72"`
	Context   UserTokenContext `json:"context" validate:"required,min=1,max=2"`
	CreatedAt *time.Time       `json:"created_at" validate:"omitempty"`
}

func (f *UserToken) Validate() error {
	if err := validate.Struct(f); err != nil {
		return errors.Join(ErrInvalidUserToken, err)
	}
	if err := f.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidUserToken, err)
	}
	if err := f.UserID.Validate(); err != nil {
		return errors.Join(ErrInvalidUserToken, err)
	}
	return nil
}

// NewUserToken creates a new UserToken.
func NewUserToken(userID ID, sentTo string, token string, context UserTokenContext) (*UserToken, error) {
	userToken := &UserToken{
		ID:      MustNewNilID(ResourceTypeUserToken),
		UserID:  userID,
		SentTo:  sentTo,
		Token:   token,
		Context: context,
	}

	if err := userToken.Validate(); err != nil {
		return nil, err
	}

	return userToken, nil
}
