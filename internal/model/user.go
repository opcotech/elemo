package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	UserIDType = "User"
)

const (
	UserStatusActive   UserStatus = iota + 1 // the user is active
	UserStatusPending                        // the user is invited but not yet active
	UserStatusInactive                       // the user is inactive
	UserStatusDeleted                        // the user is deleted
)

var (
	ErrInvalidUserDetails = errors.New("invalid user details") // the user details are invalid
	ErrInvalidUserStatus  = errors.New("invalid user status")  // the user status is invalid

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
	if s < UserStatusActive || s > UserStatusDeleted {
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

// User represents a user in the system.
type User struct {
	ID          ID         `json:"id" validate:"required,dive"`
	Username    string     `json:"username" validate:"required,lowercase,min=3,max=20,containsany=0123456789abcdefghijklmnopqrstuvwxyz-_"`
	Email       string     `json:"email" validate:"required,email"`
	Password    string     `json:"password" validate:"required,min=8"`
	Status      UserStatus `json:"status" validate:"required"`
	FirstName   string     `json:"first_name" validate:"omitempty,max=50"`
	LastName    string     `json:"last_name" validate:"omitempty,max=50"`
	Picture     string     `json:"picture" validate:"omitempty,url"`
	Title       string     `json:"title" validate:"omitempty,min=3,max=50"`
	Bio         string     `json:"bio" validate:"omitempty,min=10,max=500"`
	Phone       string     `json:"phone" validate:"omitempty,min=7,max=50"`
	Address     string     `json:"address" validate:"omitempty,min=10,max=500"`
	Links       []string   `json:"links" validate:"omitempty,dive,url"`
	Languages   []Language `json:"languages" validate:"omitempty,dive"`
	Documents   []ID       `json:"documents" validate:"omitempty,dive"`
	Permissions []ID       `json:"permissions" validate:"omitempty,dive"`
	CreatedAt   *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time `json:"updated_at" validate:"omitempty"`
}

func (u *User) Validate() error {
	if err := validate.Struct(u); err != nil {
		return errors.Join(ErrInvalidUserDetails, err)
	}
	if err := u.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidUserDetails, err)
	}
	for _, documents := range u.Documents {
		if err := documents.Validate(); err != nil {
			return errors.Join(ErrInvalidUserDetails, err)
		}
	}
	for _, permissions := range u.Permissions {
		if err := permissions.Validate(); err != nil {
			return errors.Join(ErrInvalidUserDetails, err)
		}
	}
	return nil
}

// NewUser returns a new User.
func NewUser(username, email, password string) (*User, error) {
	user := &User{
		ID:          MustNewNilID(UserIDType),
		Username:    username,
		Email:       email,
		Password:    password,
		Status:      UserStatusActive,
		Links:       make([]string, 0),
		Languages:   make([]Language, 0),
		Documents:   make([]ID, 0),
		Permissions: make([]ID, 0),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}
