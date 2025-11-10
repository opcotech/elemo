package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearNamespacesPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeNamespace.String(), pattern))
}

func clearNamespacesKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeNamespace.String(), id.String()))
}

func clearNamespacesAllGetAll(ctx context.Context, r *baseRepository) error {
	return clearNamespacesPattern(ctx, r, "GetAll", "*")
}

func clearNamespaceAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearOrganizationsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedNamespaceRepository implements caching on the
// repository.NamespaceRepository.
type CachedNamespaceRepository struct {
	cacheRepo     *baseRepository
	namespaceRepo repository.NamespaceRepository
}

func (r *CachedNamespaceRepository) Create(ctx context.Context, creatorID, orgID model.ID, namespace *model.Namespace) error {
	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearNamespaceAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.namespaceRepo.Create(ctx, creatorID, orgID, namespace)
}

func (r *CachedNamespaceRepository) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	var namespace *model.Namespace
	var err error

	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &namespace); err != nil {
		return nil, err
	}

	if namespace != nil {
		return namespace, nil
	}

	if namespace, err = r.namespaceRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *CachedNamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
	var namespaces []*model.Namespace
	var err error

	key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", orgID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &namespaces); err != nil {
		return nil, err
	}

	if namespaces != nil {
		return namespaces, nil
	}

	namespaces, err = r.namespaceRepo.GetAll(ctx, orgID, offset, limit)
	if err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (r *CachedNamespaceRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
	namespace, err := r.namespaceRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *CachedNamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearNamespacesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearNamespaceAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.namespaceRepo.Delete(ctx, id)
}

// NewCachedNamespaceRepository returns a new CachedNamespaceRepository.
func NewCachedNamespaceRepository(repo repository.NamespaceRepository, opts ...RepositoryOption) (*CachedNamespaceRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedNamespaceRepository{
		cacheRepo:     r,
		namespaceRepo: repo,
	}, nil
}
