package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearAttachmentsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeAttachment.String(), id.String()))
}

func clearAttachmentsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAttachment.String(), pattern))
}

func clearAttachmentBelongsTo(ctx context.Context, r *baseRepository, resourceID model.ID) error {
	return clearAttachmentsPattern(ctx, r, "GetAllBelongsTo", resourceID.String(), "*")
}

func clearAttachmentAllBelongsTo(ctx context.Context, r *baseRepository) error {
	return clearAttachmentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearAttachmentAllCrossCache(ctx context.Context, r *baseRepository) error {
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

// CachedAttachmentRepository implements caching on the
// repository.AttachmentRepository.
type CachedAttachmentRepository struct {
	cacheRepo      *baseRepository
	attachmentRepo repository.AttachmentRepository
}

func (r *CachedAttachmentRepository) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	if err := clearAttachmentBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}

	if err := clearAttachmentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.attachmentRepo.Create(ctx, belongsTo, attachment)
}

func (r *CachedAttachmentRepository) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	var attachment *model.Attachment
	var err error

	key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &attachment); err != nil {
		return nil, err
	}

	if attachment != nil {
		return attachment, nil
	}

	if attachment, err = r.attachmentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachment); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *CachedAttachmentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
	var attachments []*model.Attachment
	var err error

	key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &attachments); err != nil {
		return nil, err
	}

	if attachments != nil {
		return attachments, nil
	}

	if attachments, err = r.attachmentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachments); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *CachedAttachmentRepository) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	var attachment *model.Attachment
	var err error

	attachment, err = r.attachmentRepo.Update(ctx, id, name)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, attachment); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *CachedAttachmentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearAttachmentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearAttachmentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAttachmentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.attachmentRepo.Delete(ctx, id)
}

// NewCachedAttachmentRepository returns a new CachedAttachmentRepository.
func NewCachedAttachmentRepository(repo repository.AttachmentRepository, opts ...RepositoryOption) (*CachedAttachmentRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedAttachmentRepository{
		cacheRepo:      r,
		attachmentRepo: repo,
	}, nil
}
