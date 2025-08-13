package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// DocumentRepository is a repository for managing documents.
//
//go:generate mockgen -source=document.go -destination=../testutil/mock/document_repo_gen.go -package=mock -mock_names "DocumentRepository=DocumentRepository"
type DocumentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, document *model.Document) error
	Get(ctx context.Context, id model.ID) (*model.Document, error)
	GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*model.Document, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Document, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Document, error)
	Delete(ctx context.Context, id model.ID) error
}
