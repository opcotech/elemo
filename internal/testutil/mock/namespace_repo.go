package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type NamespaceRepositoryOld struct {
	mock.Mock
}

func (n *NamespaceRepositoryOld) Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error {
	args := n.Called(ctx, orgID, namespace)
	return args.Error(0)
}

func (n *NamespaceRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	args := n.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

func (n *NamespaceRepositoryOld) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
	args := n.Called(ctx, orgID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Namespace), args.Error(1)
}

func (n *NamespaceRepositoryOld) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
	args := n.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

func (n *NamespaceRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := n.Called(ctx, id)
	return args.Error(0)
}
