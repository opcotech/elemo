package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// TodoRepository is a repository for managing todos.
type TodoRepository interface {
	Create(ctx context.Context, todo *model.Todo) error
	Get(ctx context.Context, id model.ID) (*model.Todo, error)
	GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*model.Todo, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error)
	Delete(ctx context.Context, id model.ID) error
}
