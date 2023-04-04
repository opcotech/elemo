package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type UserRepository struct {
	mock.Mock
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	args := r.Called(ctx, user)
	return args.Error(0)
}

func (r *UserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
	args := r.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := r.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (r *UserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	args := r.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (r *UserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	args := r.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (r *UserRepository) Delete(ctx context.Context, id model.ID) error {
	args := r.Called(ctx, id)
	return args.Error(0)
}
