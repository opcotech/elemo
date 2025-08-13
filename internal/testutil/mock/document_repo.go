package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type DocumentRepositoryOld struct {
	mock.Mock
}

func (d *DocumentRepositoryOld) Create(ctx context.Context, belongsTo model.ID, document *model.Document) error {
	args := d.Called(ctx, belongsTo, document)
	return args.Error(0)
}

func (d *DocumentRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Document, error) {
	args := d.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (d *DocumentRepositoryOld) GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*model.Document, error) {
	args := d.Called(ctx, createdBy, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Document), args.Error(1)
}

func (d *DocumentRepositoryOld) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Document, error) {
	args := d.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Document), args.Error(1)
}

func (d *DocumentRepositoryOld) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Document, error) {
	args := d.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (d *DocumentRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := d.Called(ctx, id)
	return args.Error(0)
}
