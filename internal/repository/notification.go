package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// NotificationRepository is a repository for managing notifications.
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error)
	GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error)
	Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error)
	Delete(ctx context.Context, id, recipient model.ID) error
}
