package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type AssignmentRepositoryOld struct {
	mock.Mock
}

func (a *AssignmentRepositoryOld) Create(ctx context.Context, assignment *model.Assignment) error {
	args := a.Called(ctx, assignment)
	return args.Error(0)
}

func (a *AssignmentRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	args := a.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Assignment), args.Error(1)
}

func (a *AssignmentRepositoryOld) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
	args := a.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Assignment), args.Error(1)
}

func (a *AssignmentRepositoryOld) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
	args := a.Called(ctx, resourceID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Assignment), args.Error(1)
}

func (a *AssignmentRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := a.Called(ctx, id)
	return args.Error(0)
}
