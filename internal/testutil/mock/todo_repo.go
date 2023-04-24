package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type TodoRepository struct {
	mock.Mock
}

func (t *TodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	args := t.Called(ctx, todo)
	return args.Error(0)
}

func (t *TodoRepository) Get(ctx context.Context, id model.ID) (*model.Todo, error) {
	args := t.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (t *TodoRepository) GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*model.Todo, error) {
	args := t.Called(ctx, ownerID, offset, limit, completed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Todo), args.Error(1)
}

func (t *TodoRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error) {
	args := t.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (t *TodoRepository) Delete(ctx context.Context, id model.ID) error {
	args := t.Called(ctx, id)
	return args.Error(0)
}
