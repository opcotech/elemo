package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearLabelsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeLabel.String(), pattern))
}

func clearLabelsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeLabel.String(), id.String()))
}

func clearLabelAllGetAll(ctx context.Context, r *baseRepository) error {
	return clearLabelsPattern(ctx, r, "GetAll", "*")
}

func clearLabelAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearDocumentsPattern,
		clearIssuesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedLabelRepository implements caching on the
// repository.LabelRepository.
type CachedLabelRepository struct {
	cacheRepo *baseRepository
	labelRepo repository.LabelRepository
}

func (r *CachedLabelRepository) Create(ctx context.Context, label *model.Label) error {
	if err := clearLabelAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
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

	key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)
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
	label, err := r.labelRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
	if err := r.cacheRepo.Set(ctx, key, label); err != nil {
		return nil, err
	}

	if err := clearLabelAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return label, nil
}

func (r *CachedLabelRepository) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, labelID); err != nil {
		return err
	}

	if err := clearLabelAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.labelRepo.AttachTo(ctx, labelID, attachTo)
}

func (r *CachedLabelRepository) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, labelID); err != nil {
		return err
	}

	if err := clearLabelAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.labelRepo.DetachFrom(ctx, labelID, detachFrom)
}

func (r *CachedLabelRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := clearLabelAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
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
