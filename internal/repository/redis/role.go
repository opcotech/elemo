package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedRoleRepository implements caching on the
// repository.RoleRepository.
type CachedRoleRepository struct {
	cacheRepo *baseRepository
	roleRepo  repository.RoleRepository
}

func (r *CachedRoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	pattern := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
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

	pattern := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *CachedRoleRepository) AddMember(ctx context.Context, roleID, memberID model.ID) error {
	key := composeCacheKey(model.ResourceTypeRole.String(), roleID.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.roleRepo.AddMember(ctx, roleID, memberID)
}

func (r *CachedRoleRepository) RemoveMember(ctx context.Context, roleID, memberID model.ID) error {
	key := composeCacheKey(model.ResourceTypeRole.String(), roleID.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.roleRepo.RemoveMember(ctx, roleID, memberID)
}

func (r *CachedRoleRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
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
