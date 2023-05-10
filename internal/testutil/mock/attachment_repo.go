package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type AttachmentRepository struct {
	mock.Mock
}

func (a *AttachmentRepository) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	args := a.Called(ctx, belongsTo, attachment)
	return args.Error(0)
}

func (a *AttachmentRepository) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	args := a.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Attachment), args.Error(1)
}

func (a *AttachmentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
	args := a.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Attachment), args.Error(1)
}

func (a *AttachmentRepository) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	args := a.Called(ctx, id, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Attachment), args.Error(1)
}

func (a *AttachmentRepository) Delete(ctx context.Context, id model.ID) error {
	args := a.Called(ctx, id)
	return args.Error(0)
}
