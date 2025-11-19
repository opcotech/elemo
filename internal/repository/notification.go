package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrNotificationCreate = errors.New("failed to create notification") // the notification could not be created
	ErrNotificationDelete = errors.New("failed to delete notification") // the notification could not be deleted
	ErrNotificationRead   = errors.New("failed to read notification")   // the notification could not be retrieved
	ErrNotificationUpdate = errors.New("failed to update notification") // the notification could not be updates
)

//go:generate mockgen -source=notification.go -destination=../testutil/mock/notification_repo_gen.go -package=mock -mock_names "NotificationRepository=NotificationRepository"
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error)
	GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error)
	Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error)
	Delete(ctx context.Context, id, recipient model.ID) error
}

// NotificationRepository is a repository for managing notifications.
type PGNotificationRepository struct {
	*pgBaseRepository
}

func (r *PGNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Create")
	defer span.End()

	if err := notification.Validate(); err != nil {
		return errors.Join(ErrNotificationCreate, err)
	}

	createdAt := time.Now().UTC().Round(time.Microsecond)

	notification.ID = model.MustNewID(model.ResourceTypeNotification)
	notification.Read = false
	notification.CreatedAt = &createdAt
	notification.UpdatedAt = nil

	_, err := r.db.pool.Exec(ctx,
		"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		notification.ID, notification.Title, notification.Description, notification.Recipient,
		notification.Read, createdAt,
	)

	if err != nil {
		return errors.Join(ErrNotificationCreate, err)
	}

	return nil
}

func (r *PGNotificationRepository) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationRead, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationRead, err)
	}

	var n model.Notification
	row := r.db.pool.QueryRow(ctx, "SELECT * FROM notifications WHERE id = $1 AND recipient = $2", id, recipient)
	if err := row.Scan(&n.ID, &n.Title, &n.Description, &n.Recipient, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrNotificationRead, err)
	}

	return &n, nil
}

func (r *PGNotificationRepository) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/GetAllByRecipient")
	defer span.End()

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationRead, err)
	}

	rows, err := r.db.pool.Query(ctx,
		"SELECT * FROM notifications WHERE recipient = $1 LIMIT $2 OFFSET $3",
		recipient, limit, offset,
	)
	if err != nil {
		return nil, errors.Join(ErrNotificationRead, err)
	}
	defer rows.Close()

	notifications := make([]*model.Notification, 0)

	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.Recipient, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, errors.Join(ErrNotificationRead, err)
		}
		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (r *PGNotificationRepository) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Update")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	var n model.Notification
	row := r.db.pool.QueryRow(ctx,
		"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
		id, recipient, read,
	)
	if err := row.Scan(&n.ID, &n.Title, &n.Description, &n.Recipient, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrNotificationUpdate, err)
	}

	return &n, nil
}

func (r *PGNotificationRepository) Delete(ctx context.Context, id, recipient model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrNotificationDelete, err)
	}

	if err := recipient.Validate(); err != nil {
		return errors.Join(ErrNotificationDelete, err)
	}

	_, err := r.db.pool.Exec(ctx,
		"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
		id, recipient,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return errors.Join(ErrNotificationDelete, err)
	}

	return nil
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(opts ...PGRepositoryOption) (*PGNotificationRepository, error) {
	baseRepo, err := newPGRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &PGNotificationRepository{
		pgBaseRepository: baseRepo,
	}, nil
}

func clearNotificationsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeNotification.String(), pattern))
}

func clearNotificationsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeNotification.String(), id.String()))
}

func clearNotificationGetAllByRecipient(ctx context.Context, r *redisBaseRepository, recipient model.ID) error {
	return clearNotificationsPattern(ctx, r, "GetAllByRecipient", recipient.String(), "*")
}

// CachedNotificationRepository implements caching on the
// repository.NotificationRepository.
type RedisCachedNotificationRepository struct {
	cacheRepo        *redisBaseRepository
	notificationRepo NotificationRepository
}

func (r *RedisCachedNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if err := clearNotificationGetAllByRecipient(ctx, r.cacheRepo, notification.Recipient); err != nil {
		return err
	}

	return r.notificationRepo.Create(ctx, notification)
}

func (r *RedisCachedNotificationRepository) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
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

func (r *RedisCachedNotificationRepository) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	var err error

	key := composeCacheKey(model.ResourceTypeNotification.String(), "GetAllByRecipient", recipient.String(), offset, limit)
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

func (r *RedisCachedNotificationRepository) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "GetAllByRecipient", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return r.notificationRepo.Update(ctx, id, recipient, read)
}

func (r *RedisCachedNotificationRepository) Delete(ctx context.Context, id, recipient model.ID) error {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "GetAllByRecipient", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.notificationRepo.Delete(ctx, id, recipient)
}

// NewCachedNotificationRepository returns a new CachedNotificationRepository.
func NewCachedNotificationRepository(repo NotificationRepository, opts ...RedisRepositoryOption) (*RedisCachedNotificationRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedNotificationRepository{
		cacheRepo:        r,
		notificationRepo: repo,
	}, nil
}
