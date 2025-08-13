package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type AttachmentRepositoryOld struct {
	mock.Mock
}

func (a *AttachmentRepositoryOld) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	args := a.Called(ctx, belongsTo, attachment)
	return args.Error(0)
}

func (a *AttachmentRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	args := a.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Attachment), args.Error(1)
}

func (a *AttachmentRepositoryOld) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
	args := a.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Attachment), args.Error(1)
}

func (a *AttachmentRepositoryOld) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	args := a.Called(ctx, id, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Attachment), args.Error(1)
}

func (a *AttachmentRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := a.Called(ctx, id)
	return args.Error(0)
}
