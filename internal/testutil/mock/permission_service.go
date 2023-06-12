package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type PermissionService struct {
	mock.Mock
}

func (p *PermissionService) Create(ctx context.Context, perm *model.Permission) error {
	args := p.Called(ctx, perm)
	return args.Error(0)
}

func (p *PermissionService) CtxUserCreate(ctx context.Context, perm *model.Permission) error {
	args := p.Called(ctx, perm)
	return args.Error(0)
}

func (p *PermissionService) Get(ctx context.Context, id model.ID) (*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (p *PermissionService) GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (p *PermissionService) GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (p *PermissionService) GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	args := p.Called(ctx, source, target)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (p *PermissionService) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	args := p.Called(ctx, source, target)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionService) CtxUserHasAnyRelation(ctx context.Context, target model.ID) bool {
	args := p.Called(ctx, target)
	return args.Bool(0)
}

func (p *PermissionService) HasSystemRole(ctx context.Context, source model.ID, roles ...model.SystemRole) (bool, error) {
	args := p.Called(ctx, source, roles)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionService) CtxUserHasSystemRole(ctx context.Context, roles ...model.SystemRole) bool {
	args := p.Called(ctx, roles)
	return args.Bool(0)
}

func (p *PermissionService) HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	args := p.Called(ctx, subject, target, kinds)
	return args.Bool(0), args.Error(1)
}

func (p *PermissionService) CtxUserHasPermission(ctx context.Context, target model.ID, permissions ...model.PermissionKind) bool {
	args := p.Called(ctx, target, permissions)
	return args.Bool(0)
}

func (p *PermissionService) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	args := p.Called(ctx, id, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (p *PermissionService) Delete(ctx context.Context, id model.ID) error {
	args := p.Called(ctx, id)
	return args.Error(0)
}

func (p *PermissionService) CtxUserUpdate(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	args := p.Called(ctx, id, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (p *PermissionService) CtxUserDelete(ctx context.Context, id model.ID) error {
	args := p.Called(ctx, id)
	return args.Error(0)
}
