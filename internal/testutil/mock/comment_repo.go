package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type CommentRepository struct {
	mock.Mock
}

func (c *CommentRepository) Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error {
	args := c.Called(ctx, belongsTo, comment)
	return args.Error(0)
}

func (c *CommentRepository) Get(ctx context.Context, id model.ID) (*model.Comment, error) {
	args := c.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (c *CommentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error) {
	args := c.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Comment), args.Error(1)
}

func (c *CommentRepository) Update(ctx context.Context, id model.ID, content string) (*model.Comment, error) {
	args := c.Called(ctx, id, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (c *CommentRepository) Delete(ctx context.Context, id model.ID) error {
	args := c.Called(ctx, id)
	return args.Error(0)
}
