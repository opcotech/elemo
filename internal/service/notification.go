package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// Notification represents a notification returned by the service.
type Notification struct {
	ID          model.ID
	Title       string
	Description string
	Recipient   model.ID
	Read        bool
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateNotificationOpts holds the data required to create a notification.
type CreateNotificationOpts struct {
	Title       string   `json:"title" validate:"required,min=3,max=120"`
	Description string   `json:"description" validate:"omitempty,min=5,max=500"`
	Recipient   model.ID `json:"recipient" validate:"required"`
}

// Validate validates the create options.
func (o *CreateNotificationOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidNotificationDetails, err)
	}
	if err := o.Recipient.Validate(); err != nil {
		return model.ErrInvalidNotificationRecipient
	}
	if o.Recipient.Type != model.ResourceTypeUser {
		return model.ErrInvalidNotificationRecipient
	}
	return nil
}

// UpdateNotificationOpts holds the fields that can be updated on a notification.
type UpdateNotificationOpts struct {
	Read bool
}

// NotificationService serves the business logic of interacting with
// notifications.
//
//go:generate go tool mockgen -destination=mock/mock_notification_gen.go -package=mocksvc . NotificationService
type NotificationService interface {
	// Create creates a new notification.
	Create(ctx context.Context, opts CreateNotificationOpts) (*Notification, error)
	// Get returns a notification by its ID. If the notification does not
	// exist, an error is returned.
	Get(ctx context.Context, id, recipient model.ID) (*Notification, error)
	// ListByRecipient returns a cursor-paginated page of notifications for the
	// given recipient.
	ListByRecipient(ctx context.Context, recipient model.ID, page CursorPage) (Page[*Notification], error)
	// Update the read status of the notification. If the notification cannot
	// be updated, an error is returned.
	Update(ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts) (*Notification, error)
	// Delete deletes a notification. If the notification does not exist, an
	// error is returned.
	Delete(ctx context.Context, id, recipient model.ID) error
}

// notificationService is the concrete implementation of NotificationService.
type notificationService struct {
	runtime
	notificationRepo repository.NotificationRepository
}

func notificationFromRepository(n *repository.Notification) *Notification {
	if n == nil {
		return nil
	}
	return &Notification{
		ID:          n.ID,
		Title:       n.Title,
		Description: n.Description,
		Recipient:   n.Recipient,
		Read:        n.Read,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

// Create creates a new notification in the system.
//
// NOTE: Users should never be able to trigger notifications directly. This
// method is intended for internal (service-to-service) interactions. Exposing
// it to users through an API could lead to spams.
func (s *notificationService) Create(ctx context.Context, opts CreateNotificationOpts) (*Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Create")
	defer span.End()

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationCreate, err)
	}

	notification, err := s.notificationRepo.Create(ctx, repository.CreateNotificationOpts{
		Title:       opts.Title,
		Description: opts.Description,
		Recipient:   opts.Recipient,
	})
	if err != nil {
		return nil, errors.Join(ErrNotificationCreate, err)
	}

	return notificationFromRepository(notification), nil
}

func (s *notificationService) Get(ctx context.Context, id, recipient model.ID) (*Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Get")
	defer span.End()

	userID, err := ctxUserID(ctx)
	if err != nil || userID != recipient {
		return nil, errors.Join(ErrNotificationGet, ErrNoPermission)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	notification, err := s.notificationRepo.Get(
		ctx,
		id,
		recipient,
		repository.NotificationDetailProjection(),
	)
	if err != nil {
		return nil, errors.Join(ErrNotificationGet, err)
	}

	return notificationFromRepository(notification), nil
}

func (s *notificationService) ListByRecipient(ctx context.Context, recipient model.ID, page CursorPage) (Page[*Notification], error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/ListByRecipient")
	defer span.End()

	userID, err := ctxUserID(ctx)
	if err != nil || userID != recipient {
		return Page[*Notification]{}, errors.Join(ErrNotificationListByRecipient, ErrNoPermission)
	}

	if err := recipient.Validate(); err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationListByRecipient, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationListByRecipient, err)
	}

	notifications, err := s.notificationRepo.ListByRecipient(
		ctx,
		recipient,
		normalized,
		repository.NotificationListProjection(),
	)
	if err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationListByRecipient, err)
	}

	return mapPage(notifications, notificationFromRepository), nil
}

func (s *notificationService) Update(ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts) (*Notification, error) {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Update")
	defer span.End()

	userID, err := ctxUserID(ctx)
	if err != nil || userID != recipient {
		return nil, errors.Join(ErrNotificationUpdate, ErrNoPermission)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	notification, err := s.notificationRepo.Update(ctx, id, recipient, repository.UpdateNotificationOpts{
		Read: opts.Read,
	})
	if err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	return notificationFromRepository(notification), nil
}

func (s *notificationService) Delete(ctx context.Context, id, recipient model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.notificationService/Delete")
	defer span.End()

	userID, err := ctxUserID(ctx)
	if err != nil || userID != recipient {
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
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &notificationService{
		runtime:          rt,
		notificationRepo: notificationRepo,
	}

	if svc.notificationRepo == nil {
		return nil, ErrNoNotificationRepository
	}

	return svc, nil
}
