package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedTodoRepository is implements caching on the
// repository.TodoRepository.
type CachedTodoRepository struct {
	cacheRepo *baseRepository
	todoRepo  repository.TodoRepository
}

func (r *CachedTodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.todoRepo.Create(ctx, todo)
}

func (r *CachedTodoRepository) Get(ctx context.Context, id model.ID) (*model.Todo, error) {
	var todo *model.Todo
	var err error

	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &todo); err != nil {
		return nil, err
	}

	if todo != nil {
		return todo, nil
	}

	if todo, err = r.todoRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *CachedTodoRepository) GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*model.Todo, error) {
	var todos []*model.Todo
	var err error

	key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", ownerID.String(), offset, limit, completed)
	if err = r.cacheRepo.Get(ctx, key, &todos); err != nil {
		return nil, err
	}

	if todos != nil {
		return todos, nil
	}

	todos, err = r.todoRepo.GetByOwner(ctx, ownerID, offset, limit, completed)
	if err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, todos); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *CachedTodoRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error) {
	var todo *model.Todo
	var err error

	todo, err = r.todoRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, todo); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *CachedTodoRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.todoRepo.Delete(ctx, id)
}

// NewCachedTodoRepository returns a new CachedTodoRepository.
func NewCachedTodoRepository(repo repository.TodoRepository, opts ...RepositoryOption) (*CachedTodoRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedTodoRepository{
		cacheRepo: r,
		todoRepo:  repo,
	}, nil
}
