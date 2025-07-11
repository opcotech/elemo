package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type DocumentRepository struct {
	mock.Mock
}

func (d *DocumentRepository) Create(ctx context.Context, belongsTo model.ID, document *model.Document) error {
	args := d.Called(ctx, belongsTo, document)
	return args.Error(0)
}

func (d *DocumentRepository) Get(ctx context.Context, id model.ID) (*model.Document, error) {
	args := d.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (d *DocumentRepository) FindByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*model.Document, error) {
	args := d.Called(ctx, createdBy, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Document), args.Error(1)
}

func (d *DocumentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Document, error) {
	args := d.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Document), args.Error(1)
}

func (d *DocumentRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Document, error) {
	args := d.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (d *DocumentRepository) Delete(ctx context.Context, id model.ID) error {
	args := d.Called(ctx, id)
	return args.Error(0)
}
