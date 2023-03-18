package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	DocumentIDType = "Document"
)

var (
	ErrInvalidDocumentDetails = errors.New("invalid document details") // the document details are invalid
)

// Document represents a document in the system that can be associated with a
// model.Organization, model.Namespace, model.Project, or User. A document is a
// pointer to a file in the static file storage, editable by users with the
// appropriate permissions on the front-end.
type Document struct {
	ID        ID         `validate:"required,dive"`
	Name      string     `validate:"required,min=3,max=120"`
	Excerpt   string     `validate:"omitempty,min=10,max=500"`
	FileID    ID         `validate:"required,dive"`
	OwnedBy   ID         `validate:"required,dive"`
	CreatedAt *time.Time `validate:"omitempty"`
	UpdatedAt *time.Time `validate:"omitempty"`
}

func (d *Document) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	if err := d.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	if err := d.FileID.Validate(); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	if err := d.OwnedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	return nil
}

// NewDocument creates a new Document.
func NewDocument(name string, fileID ID, createdBy ID) (*Document, error) {
	document := &Document{
		ID:      MustNewNilID(DocumentIDType),
		Name:    name,
		FileID:  fileID,
		OwnedBy: createdBy,
	}

	if err := document.Validate(); err != nil {
		return nil, err
	}

	return document, nil
}
