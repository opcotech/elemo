package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type AssignmentRepository struct {
	mock.Mock
}

func (a *AssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	args := a.Called(ctx, assignment)
	return args.Error(0)
}

func (a *AssignmentRepository) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	args := a.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Assignment), args.Error(1)
}

func (a *AssignmentRepository) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
	args := a.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Assignment), args.Error(1)
}

func (a *AssignmentRepository) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
	args := a.Called(ctx, resourceID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Assignment), args.Error(1)
}

func (a *AssignmentRepository) Delete(ctx context.Context, id model.ID) error {
	args := a.Called(ctx, id)
	return args.Error(0)
}
