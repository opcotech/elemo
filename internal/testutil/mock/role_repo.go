package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type RoleRepository struct {
	mock.Mock
}

func (r *RoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	args := r.Called(ctx, createdBy, belongsTo, role)
	return args.Error(0)
}

func (r *RoleRepository) Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error) {
	args := r.Called(ctx, id, belongsTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (r *RoleRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	args := r.Called(ctx, belongsTo, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (r *RoleRepository) Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error) {
	args := r.Called(ctx, id, belongsTo, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (r *RoleRepository) AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	args := r.Called(ctx, roleID, memberID, belongsToID)
	return args.Error(0)
}

func (r *RoleRepository) RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	args := r.Called(ctx, roleID, memberID, belongsToID)
	return args.Error(0)
}

func (r *RoleRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	args := r.Called(ctx, id, belongsTo)
	return args.Error(0)
}
