package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedNamespaceRepository implements caching on the
// repository.NamespaceRepository.
type CachedNamespaceRepository struct {
	cacheRepo     *baseRepository
	namespaceRepo repository.NamespaceRepository
}

func (r *CachedNamespaceRepository) Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error {
	pattern := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", orgID.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.namespaceRepo.Create(ctx, orgID, namespace)
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
	var namespace *model.Namespace
	var err error

	namespace, err = r.namespaceRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *CachedNamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
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
