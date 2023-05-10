package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearProjectsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeProject.String(), pattern))
}

func clearProjectsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeProject.String(), id.String()))
}

func clearProjectsByKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return clearProjectsPattern(ctx, r, "GetByKey", id.String(), "*")
}

func clearProjectsAllGetAll(ctx context.Context, r *baseRepository) error {
	return clearProjectsPattern(ctx, r, "GetAll", "*")
}

func clearProjectsAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearNamespacesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedProjectRepository implements caching on the
// repository.ProjectRepository.
type CachedProjectRepository struct {
	cacheRepo   *baseRepository
	projectRepo repository.ProjectRepository
}

func (r *CachedProjectRepository) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
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
	if err = r.cacheRepo.Get(ctx, cacheKey, &project); err != nil {
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

	key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespaceID.String(), offset, limit)
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
	project, err := r.projectRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err := r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	if err := clearProjectsByKey(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}

	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *CachedProjectRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearProjectsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearProjectsByKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
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
