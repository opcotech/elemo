package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedAssignmentRepository implements caching on the
// repository.AssignmentRepository.
type CachedAssignmentRepository struct {
	cacheRepo      *baseRepository
	assignmentRepo repository.AssignmentRepository
}

func (r *CachedAssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	pattern := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	pattern = composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.assignmentRepo.Create(ctx, assignment)
}

func (r *CachedAssignmentRepository) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	var assignment *model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &assignment); err != nil {
		return nil, err
	}

	if assignment != nil {
		return assignment, nil
	}

	if assignment, err = r.assignmentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (r *CachedAssignmentRepository) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
	var assignments []*model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return nil, err
	}

	if assignments != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.GetByUser(ctx, userID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (r *CachedAssignmentRepository) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
	var assignments []*model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", resourceID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return nil, err
	}

	if assignments != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.GetByResource(ctx, resourceID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (r *CachedAssignmentRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	pattern = composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.assignmentRepo.Delete(ctx, id)
}

// NewCachedAssignmentRepository returns a new CachedAssignmentRepository.
func NewCachedAssignmentRepository(repo repository.AssignmentRepository, opts ...RepositoryOption) (*CachedAssignmentRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedAssignmentRepository{
		cacheRepo:      r,
		assignmentRepo: repo,
	}, nil
}
