package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// CommentRepository is a repository for managing comments.
type CommentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error
	Get(ctx context.Context, id model.ID) (*model.Comment, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error)
	Update(ctx context.Context, id model.ID, content string) (*model.Comment, error)
	Delete(ctx context.Context, id model.ID) error
}
