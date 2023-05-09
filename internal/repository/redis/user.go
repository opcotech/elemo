package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearUsersPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeUser.String(), pattern))
}

func clearUsersKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeUser.String(), id.String()))
}

func clearUsersByEmail(ctx context.Context, r *baseRepository, email string) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email))
}

func clearUsersAllByEmail(ctx context.Context, r *baseRepository) error {
	return clearUsersPattern(ctx, r, "GetByEmail", "*")
}

func clearUserAll(ctx context.Context, r *baseRepository) error {
	return clearUsersPattern(ctx, r, "GetAll", "*")
}

func clearUserAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearOrganizationsPattern,
		clearRolesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedUserRepository implements caching on the
// repository.UserRepository.
type CachedUserRepository struct {
	cacheRepo *baseRepository
	userRepo  repository.UserRepository
}

func (r *CachedUserRepository) Create(ctx context.Context, user *model.User) error {
	if err := clearUserAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearUserAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.userRepo.Create(ctx, user)
}

func (r *CachedUserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
	var user *model.User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.GetByEmail(ctx, email); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *CachedUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	var users []*model.User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &users); err != nil {
		return nil, err
	}

	if users != nil {
		return users, nil
	}

	if users, err = r.userRepo.GetAll(ctx, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *CachedUserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	var user *model.User
	var err error

	user, err = r.userRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	if err = clearUsersByEmail(ctx, r.cacheRepo, user.Email); err != nil {
		return nil, err
	}

	if err = clearUserAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *CachedUserRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearUsersKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearUsersAllByEmail(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearUserAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearUserAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.userRepo.Delete(ctx, id)
}

// NewCachedUserRepository returns a new CachedUserRepository.
func NewCachedUserRepository(repo repository.UserRepository, opts ...RepositoryOption) (*CachedUserRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedUserRepository{
		cacheRepo: r,
		userRepo:  repo,
	}, nil
}
