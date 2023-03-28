package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	AttachmentIDType = "Attachment"
)

var (
	ErrInvalidAttachmentDetails = errors.New("invalid attachment details") // the attachment details are invalid
)

// Attachment represents an attachment on a resource.
type Attachment struct {
	ID        ID         `json:"id" validate:"required,dive"`
	Name      string     `json:"name" validate:"required,min=3,max=120"`
	FileID    string     `json:"file_id" validate:"required"`
	CreatedBy ID         `json:"created_by" validate:"required,dive"`
	CreatedAt *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time `json:"updated_at" validate:"omitempty"`
}

func (c *Attachment) Validate() error {
	if err := validate.Struct(c); err != nil {
		return errors.Join(ErrInvalidAttachmentDetails, err)
	}
	if err := c.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidAttachmentDetails, err)
	}
	if err := c.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidAttachmentDetails, err)
	}
	return nil
}

// NewAttachment creates a new Attachment.
func NewAttachment(name, fileID string, createdBy ID) (*Attachment, error) {
	attachment := &Attachment{
		ID:        MustNewNilID(AttachmentIDType),
		Name:      name,
		FileID:    fileID,
		CreatedBy: createdBy,
	}

	if err := attachment.Validate(); err != nil {
		return nil, err
	}

	return attachment, nil
}
