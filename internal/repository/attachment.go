package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// AttachmentRepository is a repository for managing attachments.
type AttachmentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error
	Get(ctx context.Context, id model.ID) (*model.Attachment, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error)
	Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error)
	Delete(ctx context.Context, id model.ID) error
}
