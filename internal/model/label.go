package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	LabelIDType = "Label"
)

var (
	ErrInvalidLabelDetails = errors.New("invalid label details") // the label details are invalid
)

// Label is an entity that can be attached to a resource to provide additional
// information about it. For example, a label can be used to indicate the
// environment a resource belongs to.
type Label struct {
	ID          ID         `validate:"required,dive"`
	Name        string     `validate:"required,min=3,max=120"`
	Description string     `validate:"omitempty,min=5,max=500"`
	CreatedAt   *time.Time `validate:"omitempty"`
	UpdatedAt   *time.Time `validate:"omitempty"`
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
		ID:   MustNewNilID(LabelIDType),
		Name: name,
	}

	if err := label.Validate(); err != nil {
		return nil, err
	}

	return label, nil
}
