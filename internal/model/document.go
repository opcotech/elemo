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
	ID        ID         `json:"id" validate:"required,dive"`
	Name      string     `json:"name" validate:"required,min=3,max=120"`
	Excerpt   string     `json:"excerpt" validate:"omitempty,min=10,max=500"`
	FileID    string     `json:"file_id" validate:"required"`
	CreatedBy ID         `json:"created_by" validate:"required,dive"`
	Labels    []ID       `json:"labels" validate:"omitempty,dive"`
	Comments  []ID       `json:"comments" validate:"omitempty,dive"`
	CreatedAt *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time `json:"updated_at" validate:"omitempty"`
}

func (d *Document) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	if err := d.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	if err := d.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidDocumentDetails, err)
	}
	for _, label := range d.Labels {
		if err := label.Validate(); err != nil {
			return errors.Join(ErrInvalidDocumentDetails, err)
		}
	}
	for _, comment := range d.Comments {
		if err := comment.Validate(); err != nil {
			return errors.Join(ErrInvalidDocumentDetails, err)
		}
	}
	return nil
}

// NewDocument creates a new Document.
func NewDocument(name string, fileID string, createdBy ID) (*Document, error) {
	document := &Document{
		ID:        MustNewNilID(DocumentIDType),
		Name:      name,
		FileID:    fileID,
		CreatedBy: createdBy,
		Labels:    make([]ID, 0),
		Comments:  make([]ID, 0),
	}

	if err := document.Validate(); err != nil {
		return nil, err
	}

	return document, nil
}
