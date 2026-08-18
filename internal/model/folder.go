package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// Folder is a nested location inside an organization or namespace document
// library. Access is always evaluated against the library, never the folder.
type Folder struct {
	ID        ID         `json:"id" validate:"required"`
	Name      string     `json:"name" validate:"required,min=1,max=120"`
	CreatedBy ID         `json:"created_by" validate:"required"`
	CreatedAt *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time `json:"updated_at" validate:"omitempty"`
}

func (f *Folder) Validate() error {
	if err := validate.Struct(f); err != nil {
		return errors.Join(ErrInvalidFolderDetails, err)
	}
	if err := f.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidFolderDetails, err)
	}
	if err := f.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidFolderDetails, err)
	}
	return nil
}

// NewFolder creates a new Folder.
func NewFolder(name string, createdBy ID) (*Folder, error) {
	folder := &Folder{
		ID:        MustNewNilID(ResourceTypeFolder),
		Name:      name,
		CreatedBy: createdBy,
	}

	if err := folder.Validate(); err != nil {
		return nil, err
	}

	return folder, nil
}
