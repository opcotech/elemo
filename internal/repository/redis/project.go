package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedProjectRepository implements caching on the
// repository.ProjectRepository.
type CachedProjectRepository struct {
	cacheRepo   *baseRepository
	projectRepo repository.ProjectRepository
}

func (r *CachedProjectRepository) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	pattern := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespaceID.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.projectRepo.Create(ctx, namespaceID, project)
}

func (r *CachedProjectRepository) Get(ctx context.Context, id model.ID) (*model.Project, error) {
	var project *model.Project
	var err error

	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *CachedProjectRepository) GetByKey(ctx context.Context, key string) (*model.Project, error) {
	var project *model.Project
	var err error

	cacheKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", key)
	if err = r.cacheRepo.Get(ctx, key, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.GetByKey(ctx, key); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, cacheKey, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *CachedProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error) {
	var projects []*model.Project
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetAll", namespaceID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &projects); err != nil {
		return nil, err
	}

	if projects != nil {
		return projects, nil
	}

	if projects, err = r.projectRepo.GetAll(ctx, namespaceID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *CachedProjectRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error) {
	var project *model.Project
	var err error

	project, err = r.projectRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	pattern = composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", project.Key, "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *CachedProjectRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	pattern = composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.projectRepo.Delete(ctx, id)
}

// NewCachedProjectRepository returns a new CachedProjectRepository.
func NewCachedProjectRepository(repo repository.ProjectRepository, opts ...RepositoryOption) (*CachedProjectRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedProjectRepository{
		cacheRepo:   r,
		projectRepo: repo,
	}, nil
}
