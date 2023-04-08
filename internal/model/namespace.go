package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	NamespaceIDType = "Namespace"
)

// Namespace represents a namespace of an organization. A namespace is a
// logical grouping of Projects and Documents.
type Namespace struct {
	ID          ID         `json:"id" validate:"required,dive"`
	Name        string     `json:"name" validate:"required,min=3,max=120"`
	Description string     `json:"description" validate:"omitempty,min=5,max=500"`
	Projects    []ID       `json:"projects" validate:"omitempty,dive"`
	Documents   []ID       `json:"documents" validate:"omitempty,dive"`
	CreatedAt   *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time `json:"updated_at" validate:"omitempty"`
}

func (n *Namespace) Validate() error {
	if err := validate.Struct(n); err != nil {
		return errors.Join(ErrInvalidNamespaceDetails, err)
	}
	if err := n.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidNamespaceDetails, err)
	}
	return nil
}

// NewNamespace creates a new Namespace.
func NewNamespace(name string) (*Namespace, error) {
	namespace := &Namespace{
		ID:        MustNewNilID(NamespaceIDType),
		Name:      name,
		Projects:  make([]ID, 0),
		Documents: make([]ID, 0),
	}

	if err := namespace.Validate(); err != nil {
		return nil, err
	}

	return namespace, nil
}
