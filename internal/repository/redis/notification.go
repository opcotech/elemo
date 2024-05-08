package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearNotificationsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeNotification.String(), pattern))
}

func clearNotificationsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeNotification.String(), id.String()))
}

func clearNotificationGetByRecipient(ctx context.Context, r *baseRepository, recipient model.ID) error {
	return clearNotificationsPattern(ctx, r, "GetByRecipient", recipient.String(), "*")
}

// CachedNotificationRepository implements caching on the
// repository.NotificationRepository.
type CachedNotificationRepository struct {
	cacheRepo        *baseRepository
	notificationRepo repository.NotificationRepository
}

func (r *CachedNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if err := clearNotificationGetByRecipient(ctx, r.cacheRepo, notification.Recipient); err != nil {
		return err
	}

	return r.notificationRepo.Create(ctx, notification)
}

func (r *CachedNotificationRepository) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
	var notification *model.Notification
	var err error

	key := composeCacheKey(model.ResourceTypeNotification.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &notification); err != nil {
		return nil, err
	}

	if notification != nil {
		return notification, nil
	}

	if notification, err = r.notificationRepo.Get(ctx, id, recipient); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (r *CachedNotificationRepository) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	var err error

	key := composeCacheKey(model.ResourceTypeNotification.String(), "GetByRecipient", recipient.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &notifications); err != nil {
		return nil, err
	}

	if notifications != nil {
		return notifications, nil
	}

	notifications, err = r.notificationRepo.GetAllByRecipient(ctx, recipient, offset, limit)
	if err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *CachedNotificationRepository) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "GetByRecipient", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return r.notificationRepo.Update(ctx, id, recipient, read)
}

func (r *CachedNotificationRepository) Delete(ctx context.Context, id, recipient model.ID) error {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "GetByRecipient", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.notificationRepo.Delete(ctx, id, recipient)
}

// NewCachedNotificationRepository returns a new CachedNotificationRepository.
func NewCachedNotificationRepository(repo repository.NotificationRepository, opts ...RepositoryOption) (*CachedNotificationRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedNotificationRepository{
		cacheRepo:        r,
		notificationRepo: repo,
	}, nil
}
