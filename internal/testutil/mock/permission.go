package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type PermissionRepository struct {
	mock.Mock
}

func (p *PermissionRepository) Create(ctx context.Context, perm *model.Permission) error {
	args := p.Called(ctx, perm)
	return args.Error(0)
}

func (p *PermissionRepository) Get(ctx context.Context, id model.ID) (*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (p *PermissionRepository) GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (p *PermissionRepository) GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (p *PermissionRepository) HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	args := p.Called(ctx, subject, target, kinds)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionRepository) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	args := p.Called(ctx, source, target)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionRepository) HasSystemRole(ctx context.Context, source model.ID, targets ...model.SystemRole) (bool, error) {
	args := p.Called(ctx, source, targets)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionRepository) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	args := p.Called(ctx, id, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (p *PermissionRepository) Delete(ctx context.Context, id model.ID) error {
	args := p.Called(ctx, id)
	return args.Error(0)
}
