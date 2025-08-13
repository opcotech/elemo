package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// AttachmentRepository is a repository for managing attachments.
//
//go:generate mockgen -source=attachment.go -destination=../testutil/mock/attachment_repo_gen.go -package=mock -mock_names "AttachmentRepository=AttachmentRepository"
type AttachmentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error
	Get(ctx context.Context, id model.ID) (*model.Attachment, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error)
	Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error)
	Delete(ctx context.Context, id model.ID) error
}
