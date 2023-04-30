package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearAssignmentsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAssignment.String(), pattern))
}

func clearAssignmentsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeAssignment.String(), id.String()))
}

func clearAssignmentByResource(ctx context.Context, r *baseRepository, resourceID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "GetByResource", resourceID.String(), "*")
}

func clearAssignmentAllByResource(ctx context.Context, r *baseRepository) error {
	return clearAssignmentsPattern(ctx, r, "GetByResource", "*")
}

func clearAssignmentByUser(ctx context.Context, r *baseRepository, userID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "GetByUser", userID.String(), "*")
}

func clearAssignmentAllByUser(ctx context.Context, r *baseRepository) error {
	return clearAssignmentsPattern(ctx, r, "GetByUser", "*")
}

func clearAssignmentAllCrossCache(ctx context.Context, r *baseRepository, assignment *model.Assignment) error {
	deleteFns := make([]func(context.Context, *baseRepository, ...string) error, 0)

	if assignment == nil {
		deleteFns = append(deleteFns, clearDocumentsPattern, clearIssuesPattern)
	}

	if assignment != nil {
		switch assignment.Resource.Type {
		case model.ResourceTypeDocument:
			deleteFns = append(deleteFns, clearDocumentsPattern)
		case model.ResourceTypeIssue:
			deleteFns = append(deleteFns, clearIssuesPattern)
		default:
			return ErrUnexpectedCachedResource
		}
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedAssignmentRepository implements caching on the
// repository.AssignmentRepository.
type CachedAssignmentRepository struct {
	cacheRepo      *baseRepository
	assignmentRepo repository.AssignmentRepository
}

func (r *CachedAssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	if err := clearAssignmentByResource(ctx, r.cacheRepo, assignment.Resource); err != nil {
		return err
	}

	if err := clearAssignmentByUser(ctx, r.cacheRepo, assignment.User); err != nil {
		return err
	}

	if err := clearAssignmentAllCrossCache(ctx, r.cacheRepo, assignment); err != nil {
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
	if err := clearAssignmentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearAssignmentAllByResource(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAssignmentAllByUser(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAssignmentAllCrossCache(ctx, r.cacheRepo, nil); err != nil {
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
