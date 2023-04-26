package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// NamespaceRepository is a repository for managing namespaces.
type NamespaceRepository interface {
	Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error
	Get(ctx context.Context, id model.ID) (*model.Namespace, error)
	GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error)
	Delete(ctx context.Context, id model.ID) error
}
