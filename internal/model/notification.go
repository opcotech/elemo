package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// Notification represents an in-app notification that is usually sent by the
// application.
type Notification struct {
	ID          ID         `json:"id" validate:"required"`
	Title       string     `json:"name" validate:"required,min=3,max=120"`
	Description string     `json:"description" validate:"omitempty,min=5,max=500"`
	Recipient   ID         `json:"recipient" validate:"required"`
	Read        bool       `json:"read"`
	CreatedAt   *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time `json:"updated_at" validate:"omitempty"`
}

func (n *Notification) Validate() error {
	if err := validate.Struct(n); err != nil {
		return errors.Join(ErrInvalidNotificationDetails, err)
	}
	if err := n.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidNotificationDetails, err)
	}
	if err := n.Recipient.Validate(); err != nil {
		return ErrInvalidNotificationRecipient
	}
	if n.Recipient.Type != ResourceTypeUser {
		return ErrInvalidNotificationRecipient
	}
	return nil
}

// NewNotification creates a new Notification.
func NewNotification(title string, recipient ID) (*Notification, error) {
	notification := &Notification{
		ID:        MustNewNilID(ResourceTypeNotification),
		Title:     title,
		Recipient: recipient,
		Read:      false,
	}

	if err := notification.Validate(); err != nil {
		return nil, err
	}

	return notification, nil
}
