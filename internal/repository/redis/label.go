package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// CachedLabelRepository implements caching on the
// repository.LabelRepository.
type CachedLabelRepository struct {
	cacheRepo *baseRepository
	labelRepo repository.LabelRepository
}

func (r *CachedLabelRepository) Create(ctx context.Context, label *model.Label) error {
	pattern := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.labelRepo.Create(ctx, label)
}

func (r *CachedLabelRepository) Get(ctx context.Context, id model.ID) (*model.Label, error) {
	var label *model.Label
	var err error

	key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &label); err != nil {
		return nil, err
	}

	if label != nil {
		return label, nil
	}

	if label, err = r.labelRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (r *CachedLabelRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.Label, error) {
	var labels []*model.Label
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetAll", offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &labels); err != nil {
		return nil, err
	}

	if labels != nil {
		return labels, nil
	}

	if labels, err = r.labelRepo.GetAll(ctx, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, labels); err != nil {
		return nil, err
	}

	return labels, nil
}

func (r *CachedLabelRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Label, error) {
	var label *model.Label
	var err error

	label, err = r.labelRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, label); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return label, nil
}

func (r *CachedLabelRepository) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	key := composeCacheKey(model.ResourceTypeLabel.String(), labelID.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.labelRepo.AttachTo(ctx, labelID, attachTo)
}

func (r *CachedLabelRepository) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	key := composeCacheKey(model.ResourceTypeLabel.String(), labelID.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.labelRepo.DetachFrom(ctx, labelID, detachFrom)
}

func (r *CachedLabelRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.labelRepo.Delete(ctx, id)
}

// NewCachedLabelRepository returns a new CachedLabelRepository.
func NewCachedLabelRepository(repo repository.LabelRepository, opts ...RepositoryOption) (*CachedLabelRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedLabelRepository{
		cacheRepo: r,
		labelRepo: repo,
	}, nil
}
