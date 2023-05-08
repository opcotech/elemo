package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearRolesPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeRole.String(), pattern))
}

func clearRolesKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeRole.String(), id.String()))
}

func clearRolesBelongsTo(ctx context.Context, r *baseRepository, id model.ID) error {
	return clearRolesPattern(ctx, r, "GetAllBelongsTo", id.String(), "*")
}

func clearRolesAllBelongsTo(ctx context.Context, r *baseRepository) error {
	return clearRolesPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearRoleAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearOrganizationsPattern,
		clearProjectsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedRoleRepository implements caching on the
// repository.RoleRepository.
type CachedRoleRepository struct {
	cacheRepo *baseRepository
	roleRepo  repository.RoleRepository
}

func (r *CachedRoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	if err := clearRolesBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}

	return r.roleRepo.Create(ctx, createdBy, belongsTo, role)
}

func (r *CachedRoleRepository) Get(ctx context.Context, id model.ID) (*model.Role, error) {
	var role *model.Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &role); err != nil {
		return nil, err
	}

	if role != nil {
		return role, nil
	}

	if role, err = r.roleRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *CachedRoleRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	var roles []*model.Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &roles); err != nil {
		return nil, err
	}

	if roles != nil {
		return roles, nil
	}

	if roles, err = r.roleRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *CachedRoleRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Role, error) {
	var role *model.Role
	var err error

	role, err = r.roleRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *CachedRoleRepository) AddMember(ctx context.Context, roleID, memberID model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, roleID); err != nil {
		return err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.roleRepo.AddMember(ctx, roleID, memberID)
}

func (r *CachedRoleRepository) RemoveMember(ctx context.Context, roleID, memberID model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, roleID); err != nil {
		return err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.roleRepo.RemoveMember(ctx, roleID, memberID)
}

func (r *CachedRoleRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearRoleAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.roleRepo.Delete(ctx, id)
}

// NewCachedRoleRepository returns a new CachedRoleRepository.
func NewCachedRoleRepository(repo repository.RoleRepository, opts ...RepositoryOption) (*CachedRoleRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedRoleRepository{
		cacheRepo: r,
		roleRepo:  repo,
	}, nil
}
