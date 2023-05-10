package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// LabelRepository is a repository for managing labels.
type LabelRepository interface {
	Create(ctx context.Context, label *model.Label) error
	Get(ctx context.Context, id model.ID) (*model.Label, error)
	GetAll(ctx context.Context, offset, limit int) ([]*model.Label, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Label, error)
	AttachTo(ctx context.Context, labelID, attachTo model.ID) error
	DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error
	Delete(ctx context.Context, id model.ID) error
}
