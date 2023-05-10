package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type NamespaceRepository struct {
	mock.Mock
}

func (n *NamespaceRepository) Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error {
	args := n.Called(ctx, orgID, namespace)
	return args.Error(0)
}

func (n *NamespaceRepository) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	args := n.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

func (n *NamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
	args := n.Called(ctx, orgID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Namespace), args.Error(1)
}

func (n *NamespaceRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
	args := n.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

func (n *NamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	args := n.Called(ctx, id)
	return args.Error(0)
}
