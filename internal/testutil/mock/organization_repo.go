package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type OrganizationRepository struct {
	mock.Mock
}

func (o *OrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	args := o.Called(ctx, owner, organization)
	return args.Error(0)
}

func (o *OrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	args := o.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (o *OrganizationRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	args := o.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Organization), args.Error(1)
}

func (o *OrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	args := o.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (o *OrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	args := o.Called(ctx, orgID, memberID)
	return args.Error(0)
}

func (o *OrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	args := o.Called(ctx, orgID, memberID)
	return args.Error(0)
}

func (o *OrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	args := o.Called(ctx, id)
	return args.Error(0)
}
