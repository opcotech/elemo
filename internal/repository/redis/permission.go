package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearPermissionAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearRolesPattern,
		clearUsersPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedPermissionRepository implements cache clearing on resources that are
// related dependent on permission changes. This repository does not cache any
// data to prevent stale permission data. This repository mostly acts as a
// proxy to the permission repository and clears the cache on any changes.
//
// Adding permission caching could be a future improvement, but that's a
// double-edged sword. It would be a performance improvement, but it would also
// mean that stale data could be cached.
type CachedPermissionRepository struct {
	cacheRepo      *baseRepository
	permissionRepo repository.PermissionRepository
}

func (c *CachedPermissionRepository) Create(ctx context.Context, perm *model.Permission) error {
	if err := clearPermissionAllCrossCache(ctx, c.cacheRepo); err != nil {
		return err
	}
	return c.permissionRepo.Create(ctx, perm)
}

func (c *CachedPermissionRepository) Get(ctx context.Context, id model.ID) (*model.Permission, error) {
	return c.permissionRepo.Get(ctx, id)
}

func (c *CachedPermissionRepository) GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	return c.permissionRepo.GetBySubject(ctx, id)
}

func (c *CachedPermissionRepository) GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	return c.permissionRepo.GetByTarget(ctx, id)
}

func (c *CachedPermissionRepository) GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	return c.permissionRepo.GetBySubjectAndTarget(ctx, source, target)
}

func (c *CachedPermissionRepository) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	if err := clearPermissionAllCrossCache(ctx, c.cacheRepo); err != nil {
		return nil, err
	}
	return c.permissionRepo.Update(ctx, id, kind)
}

func (c *CachedPermissionRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearPermissionAllCrossCache(ctx, c.cacheRepo); err != nil {
		return err
	}
	return c.permissionRepo.Delete(ctx, id)
}

func (c *CachedPermissionRepository) HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	return c.permissionRepo.HasPermission(ctx, subject, target, kinds...)
}

func (c *CachedPermissionRepository) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	return c.permissionRepo.HasAnyRelation(ctx, source, target)
}

func (c *CachedPermissionRepository) HasSystemRole(ctx context.Context, source model.ID, roles ...model.SystemRole) (bool, error) {
	return c.permissionRepo.HasSystemRole(ctx, source, roles...)
}

// NewCachedPermissionRepository returns a new CachedPermissionRepository.
func NewCachedPermissionRepository(repo repository.PermissionRepository, opts ...RepositoryOption) (*CachedPermissionRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedPermissionRepository{
		cacheRepo:      r,
		permissionRepo: repo,
	}, nil
}
