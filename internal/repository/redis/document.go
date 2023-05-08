package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearDocumentsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeDocument.String(), pattern))
}

func clearDocumentsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeDocument.String(), id.String()))
}

func clearDocumentBelongsTo(ctx context.Context, r *baseRepository, belongsToID model.ID) error {
	return clearDocumentsPattern(ctx, r, "GetAllBelongsTo", belongsToID.String(), "*")
}

func clearDocumentAllBelongsTo(ctx context.Context, r *baseRepository) error {
	return clearDocumentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearDocumentByCreator(ctx context.Context, r *baseRepository, createdByID model.ID) error {
	return clearDocumentsPattern(ctx, r, "GetByCreator", createdByID.String(), "*")
}

func clearDocumentAllByCreator(ctx context.Context, r *baseRepository) error {
	return clearDocumentsPattern(ctx, r, "GetByCreator", "*")
}

func clearDocumentAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearNamespacesPattern,
		clearProjectsPattern,
		clearUsersPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedDocumentRepository implements caching on the
// repository.DocumentRepository.
type CachedDocumentRepository struct {
	cacheRepo    *baseRepository
	documentRepo repository.DocumentRepository
}

func (r *CachedDocumentRepository) Create(ctx context.Context, belongsTo model.ID, document *model.Document) error {
	if err := clearDocumentBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}

	if err := clearDocumentByCreator(ctx, r.cacheRepo, document.CreatedBy); err != nil {
		return err
	}

	if err := clearDocumentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.documentRepo.Create(ctx, belongsTo, document)
}

func (r *CachedDocumentRepository) Get(ctx context.Context, id model.ID) (*model.Document, error) {
	var document *model.Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &document); err != nil {
		return nil, err
	}

	if document != nil {
		return document, nil
	}

	if document, err = r.documentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	return document, nil
}

func (r *CachedDocumentRepository) GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*model.Document, error) {
	var documents []*model.Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return nil, err
	}

	if documents != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.GetByCreator(ctx, createdBy, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *CachedDocumentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Document, error) {
	var documents []*model.Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return nil, err
	}

	if documents != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *CachedDocumentRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Document, error) {
	var document *model.Document
	var err error

	document, err = r.documentRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	if err := clearDocumentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if err = clearDocumentByCreator(ctx, r.cacheRepo, document.CreatedBy); err != nil {
		return nil, err
	}

	return document, nil
}

func (r *CachedDocumentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearDocumentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearDocumentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearDocumentAllByCreator(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearDocumentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.documentRepo.Delete(ctx, id)
}

// NewCachedDocumentRepository returns a new CachedDocumentRepository.
func NewCachedDocumentRepository(repo repository.DocumentRepository, opts ...RepositoryOption) (*CachedDocumentRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedDocumentRepository{
		cacheRepo:    r,
		documentRepo: repo,
	}, nil
}
