package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// ProjectRepository is a repository for managing projects.
type ProjectRepository interface {
	Create(ctx context.Context, namespaceID model.ID, project *model.Project) error
	Get(ctx context.Context, id model.ID) (*model.Project, error)
	GetByKey(ctx context.Context, key string) (*model.Project, error)
	GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error)
	Delete(ctx context.Context, id model.ID) error
}
