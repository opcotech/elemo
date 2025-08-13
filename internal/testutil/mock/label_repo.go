package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type LabelRepositoryOld struct {
	mock.Mock
}

func (l *LabelRepositoryOld) Create(ctx context.Context, label *model.Label) error {
	args := l.Called(ctx, label)
	return args.Error(0)
}

func (l *LabelRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Label, error) {
	args := l.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Label), args.Error(1)
}

func (l *LabelRepositoryOld) GetAll(ctx context.Context, offset, limit int) ([]*model.Label, error) {
	args := l.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Label), args.Error(1)
}

func (l *LabelRepositoryOld) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Label, error) {
	args := l.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Label), args.Error(1)
}

func (l *LabelRepositoryOld) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	args := l.Called(ctx, labelID, attachTo)
	return args.Error(0)
}

func (l *LabelRepositoryOld) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	args := l.Called(ctx, labelID, detachFrom)
	return args.Error(0)
}

func (l *LabelRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := l.Called(ctx, id)
	return args.Error(0)
}
