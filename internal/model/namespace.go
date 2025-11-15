package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// NamespaceProject represents a simplified project response within a namespace.
type NamespaceProject struct {
	ID          ID            `json:"id" validate:"required"`
	Key         string        `json:"key" validate:"required,alpha,min=3,max=6"`
	Name        string        `json:"name" validate:"required,min=3,max=120"`
	Description string        `json:"description" validate:"omitempty,min=10,max=500"`
	Logo        string        `json:"logo" validate:"omitempty,url"`
	Status      ProjectStatus `json:"status" validate:"required,min=1,max=2"`
}

// Validate validates the namespace project details.
func (np *NamespaceProject) Validate() error {
	if err := validate.Struct(np); err != nil {
		return errors.Join(ErrInvalidNamespaceProjectDetails, err)
	}
	if err := np.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidNamespaceProjectDetails, err)
	}
	if np.ID.Type != ResourceTypeProject {
		return errors.Join(ErrInvalidNamespaceProjectDetails, ErrInvalidResourceType)
	}
	return nil
}

// NewNamespaceProject creates a new NamespaceProject.
func NewNamespaceProject(id ID, key, name string, description, logo string, status ProjectStatus) (*NamespaceProject, error) {
	project := &NamespaceProject{
		ID:          id,
		Key:         key,
		Name:        name,
		Description: description,
		Logo:        logo,
		Status:      status,
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	return project, nil
}

// NamespaceDocument represents a simplified document response within a namespace.
type NamespaceDocument struct {
	ID        ID         `json:"id" validate:"required"`
	Name      string     `json:"name" validate:"required,min=3,max=120"`
	Excerpt   string     `json:"excerpt" validate:"omitempty,min=10,max=500"`
	CreatedBy ID         `json:"created_by" validate:"required"`
	CreatedAt *time.Time `json:"created_at" validate:"omitempty"`
}

// Validate validates the namespace document details.
func (nd *NamespaceDocument) Validate() error {
	if err := validate.Struct(nd); err != nil {
		return errors.Join(ErrInvalidNamespaceDocumentDetails, err)
	}
	if err := nd.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidNamespaceDocumentDetails, err)
	}
	if nd.ID.Type != ResourceTypeDocument {
		return errors.Join(ErrInvalidNamespaceDocumentDetails, ErrInvalidResourceType)
	}
	if err := nd.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidNamespaceDocumentDetails, err)
	}
	return nil
}

// NewNamespaceDocument creates a new NamespaceDocument.
func NewNamespaceDocument(id ID, name, excerpt string, createdBy ID, createdAt *time.Time) (*NamespaceDocument, error) {
	document := &NamespaceDocument{
		ID:        id,
		Name:      name,
		Excerpt:   excerpt,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
	}

	if err := document.Validate(); err != nil {
		return nil, err
	}

	return document, nil
}

// Namespace represents a namespace of an organization. A namespace is a
// logical grouping of Projects and Documents.
type Namespace struct {
	ID          ID                   `json:"id" validate:"required"`
	Name        string               `json:"name" validate:"required,min=3,max=120"`
	Description string               `json:"description" validate:"omitempty,min=5,max=500"`
	Projects    []*NamespaceProject  `json:"projects" validate:"omitempty,dive"`
	Documents   []*NamespaceDocument `json:"documents" validate:"omitempty,dive"`
	CreatedAt   *time.Time           `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time           `json:"updated_at" validate:"omitempty"`
}

func (n *Namespace) Validate() error {
	if err := validate.Struct(n); err != nil {
		return errors.Join(ErrInvalidNamespaceDetails, err)
	}
	if err := n.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidNamespaceDetails, err)
	}
	for _, project := range n.Projects {
		if err := project.Validate(); err != nil {
			return errors.Join(ErrInvalidNamespaceDetails, err)
		}
	}
	for _, document := range n.Documents {
		if err := document.Validate(); err != nil {
			return errors.Join(ErrInvalidNamespaceDetails, err)
		}
	}
	return nil
}

// NewNamespace creates a new Namespace.
func NewNamespace(name string) (*Namespace, error) {
	namespace := &Namespace{
		ID:        MustNewNilID(ResourceTypeNamespace),
		Name:      name,
		Projects:  make([]*NamespaceProject, 0),
		Documents: make([]*NamespaceDocument, 0),
	}

	if err := namespace.Validate(); err != nil {
		return nil, err
	}

	return namespace, nil
}
