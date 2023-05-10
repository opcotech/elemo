package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearCommentsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeComment.String(), id.String()))
}

func clearCommentsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeComment.String(), pattern))
}

func clearCommentBelongsTo(ctx context.Context, r *baseRepository, resourceID model.ID) error {
	switch resourceID.Type {
	case model.ResourceTypeDocument:
		if err := clearDocumentsPattern(ctx, r, "*"); err != nil {
			return err
		}
	case model.ResourceTypeIssue:
		if err := clearIssuesPattern(ctx, r, "*"); err != nil {
			return err
		}
	}

	return clearCommentsPattern(ctx, r, "GetAllBelongsTo", resourceID.String(), "*")
}

func clearCommentAllBelongsTo(ctx context.Context, r *baseRepository) error {
	return clearCommentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearCommentAllCrossCache(ctx context.Context, r *baseRepository) error {
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

// CachedCommentRepository implements caching on the
// repository.CommentRepository.
type CachedCommentRepository struct {
	cacheRepo   *baseRepository
	commentRepo repository.CommentRepository
}

func (r *CachedCommentRepository) Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error {
	if err := clearCommentBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}
	return r.commentRepo.Create(ctx, belongsTo, comment)
}

func (r *CachedCommentRepository) Get(ctx context.Context, id model.ID) (*model.Comment, error) {
	var comment *model.Comment
	var err error

	key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &comment); err != nil {
		return nil, err
	}

	if comment != nil {
		return comment, nil
	}

	if comment, err = r.commentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *CachedCommentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error) {
	var comments []*model.Comment
	var err error

	key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &comments); err != nil {
		return nil, err
	}

	if comments != nil {
		return comments, nil
	}

	if comments, err = r.commentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *CachedCommentRepository) Update(ctx context.Context, id model.ID, content string) (*model.Comment, error) {
	var comment *model.Comment
	var err error

	comment, err = r.commentRepo.Update(ctx, id, content)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, comment); err != nil {
		return nil, err
	}

	if err := clearCommentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *CachedCommentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearCommentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearCommentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearCommentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.commentRepo.Delete(ctx, id)
}

// NewCachedCommentRepository returns a new CachedCommentRepository.
func NewCachedCommentRepository(repo repository.CommentRepository, opts ...RepositoryOption) (*CachedCommentRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedCommentRepository{
		cacheRepo:   r,
		commentRepo: repo,
	}, nil
}
