package mock

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/stretchr/testify/mock"
)

type NotificationRepositoryOld struct {
	mock.Mock
}

func (n *NotificationRepositoryOld) Create(ctx context.Context, notification *model.Notification) error {
	args := n.Called(ctx, notification)
	return args.Error(0)
}

func (n *NotificationRepositoryOld) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
	args := n.Called(ctx, id, recipient)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), nil
}

func (n *NotificationRepositoryOld) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	args := n.Called(ctx, recipient, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), nil
}

func (n *NotificationRepositoryOld) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	args := n.Called(ctx, id, recipient, read)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), nil
}

func (n *NotificationRepositoryOld) Delete(ctx context.Context, id, recipient model.ID) error {
	args := n.Called(ctx, id, recipient)
	return args.Error(0)
}
