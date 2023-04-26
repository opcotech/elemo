package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// Label is an entity that can be attached to a resource to provide additional
// information about it. For example, a Type can be used to indicate the
// environment a resource belongs to.
type Label struct {
	ID          ID         `json:"id" validate:"required,dive"`
	Name        string     `json:"name" validate:"required,min=3,max=120"`
	Description string     `json:"description" validate:"omitempty,min=5,max=500"`
	CreatedAt   *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time `json:"updated_at" validate:"omitempty"`
}

func (l *Label) Validate() error {
	if err := validate.Struct(l); err != nil {
		return errors.Join(ErrInvalidLabelDetails, err)
	}
	if err := l.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidLabelDetails, err)
	}
	return nil
}

// NewLabel creates a new Label.
func NewLabel(name string) (*Label, error) {
	label := &Label{
		ID:   MustNewNilID(ResourceTypeLabel),
		Name: name,
	}

	if err := label.Validate(); err != nil {
		return nil, err
	}

	return label, nil
}
