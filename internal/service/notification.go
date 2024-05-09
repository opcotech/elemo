package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

// NotificationService serves the business logic of interacting with
// notifications.
type NotificationService interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *model.Notification) error
	// Get returns an notification by its ID. If the notification does not
	// exist, an error is returned.
	Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error)
	// GetAllByRecipient returns all notifications for the given recipient. The
	// offset and limit parameters are  used to paginate the results. If the
	// offset is greater than the number of notification in the system, an empty
	// slice is returned.
	GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error)
	// Update the read status of the notification. If the notification cannot be
	// updated, an error is returned.
	Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error)
	// Delete deletes an notification. If the notification does not exist, an
	// error is returned.
	Delete(ctx context.Context, id, recipient model.ID) error
}

// notificationService is the concrete implementation of NotificationService.
type notificationService struct {
	*baseService
	notificationRepo repository.NotificationRepository
}

// Create creates a new notification in the system.
//
// NOTE: Users should never be able to trigger notifications directly. This
// method is intended for internal (service-to-service) interactions. Exposing
// it to users through an API could lead to spams.
func (s *notificationService) Create(ctx context.Context, notification *model.Notification) error {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Create")
	defer span.End()

	if err := notification.Validate(); err != nil {
		return errors.Join(ErrNotificationCreate, err)
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return errors.Join(ErrNotificationCreate, err)
	}

	return nil
}

func (s *notificationService) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Get")
	defer span.End()

	if userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok || userID != recipient {
		return nil, errors.Join(ErrNotificationGet, ErrNoPermission)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	notification, err := s.notificationRepo.Get(ctx, id, recipient)
	if err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	return notification, nil
}

func (s *notificationService) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/GetAllByRecipient")
	defer span.End()

	if userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok || userID != recipient {
		return nil, errors.Join(ErrNotificationGetAllByRecipient, ErrNoPermission)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationGetAllByRecipient, err)
	}

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrNotificationGetAllByRecipient, ErrInvalidPaginationParams)
	}

	notifications, err := s.notificationRepo.GetAllByRecipient(ctx, recipient, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrNotificationGetAllByRecipient, err)
	}

	return notifications, nil
}

func (s *notificationService) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Update")
	defer span.End()

	if userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok || userID != recipient {
		return nil, errors.Join(ErrNotificationUpdate, ErrNoPermission)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	notification, err := s.notificationRepo.Update(ctx, id, recipient, read)
	if err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	return notification, nil
}

func (s *notificationService) Delete(ctx context.Context, id, recipient model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Delete")
	defer span.End()

	if userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok || userID != recipient {
		return errors.Join(ErrNotificationDelete, ErrNoPermission)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrNotificationDelete, err)
	}

	if err := recipient.Validate(); err != nil {
		return errors.Join(ErrNotificationDelete, err)
	}

	if err := s.notificationRepo.Delete(ctx, id, recipient); err != nil {
		return errors.Join(ErrNotificationDelete, err)
	}

	return nil
}

// NewNotificationService returns a new instance of the NotificationService
// interface.
func NewNotificationService(notificationRepo repository.NotificationRepository, opts ...Option) (NotificationService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &notificationService{
		baseService:      s,
		notificationRepo: notificationRepo,
	}

	if svc.notificationRepo == nil {
		return nil, ErrNoNotificationRepository
	}

	return svc, nil
}
