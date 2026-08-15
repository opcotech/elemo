package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrNotificationCreate = errors.New("failed to create notification") // the notification could not be created
	ErrNotificationDelete = errors.New("failed to delete notification") // the notification could not be deleted
	ErrNotificationRead   = errors.New("failed to read notification")   // the notification could not be retrieved
	ErrNotificationUpdate = errors.New("failed to update notification") // the notification could not be updates
)

// Notification represents a notification persisted by the repository.
type Notification struct {
	ID          model.ID   `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Recipient   model.ID   `json:"recipient"`
	Read        bool       `json:"read"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// CreateNotificationOpts holds the data required to create a notification.
type CreateNotificationOpts struct {
	Title       string
	Description string
	Recipient   model.ID
}

// UpdateNotificationOpts holds the fields that can be updated on a notification.
type UpdateNotificationOpts struct {
	Read bool
}

//go:generate go tool mockgen -source=notification.go -destination=notification_mock_gen.go -package=repository -mock_names "NotificationRepository=MockNotificationRepository"
type NotificationRepository interface {
	Create(ctx context.Context, opts CreateNotificationOpts) (*Notification, error)
	Get(ctx context.Context, id, recipient model.ID, proj NotificationProjection) (*Notification, error)
	ListByRecipient(ctx context.Context, recipient model.ID, page CursorPage, proj NotificationProjection) (Page[*Notification], error)
	Update(ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts) (*Notification, error)
	Delete(ctx context.Context, id, recipient model.ID) error
}

// PGNotificationRepository is a repository for managing notifications.
type PGNotificationRepository struct {
	*pgBaseRepository
}

func (r *PGNotificationRepository) Create(ctx context.Context, opts CreateNotificationOpts) (*Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Create")
	defer span.End()

	notification := &Notification{
		ID:          model.MustNewID(model.ResourceTypeNotification),
		Title:       opts.Title,
		Description: opts.Description,
		Recipient:   opts.Recipient,
		Read:        false,
		CreatedAt:   convert.ToPointer(time.Now().UTC().Round(time.Microsecond)),
		UpdatedAt:   nil,
	}

	_, err := r.db.pool.Exec(ctx,
		"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		notification.ID, notification.Title, notification.Description, notification.Recipient,
		notification.Read, *notification.CreatedAt,
	)
	if err != nil {
		return nil, errors.Join(ErrNotificationCreate, err)
	}

	return notification, nil
}

func (r *PGNotificationRepository) Get(ctx context.Context, id, recipient model.ID, _ NotificationProjection) (*Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Get")
	defer span.End()

	row := r.db.pool.QueryRow(ctx, "SELECT * FROM notifications WHERE id = $1 AND recipient = $2", id, recipient)
	n, err := scanNotificationRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrNotificationRead, err)
	}

	return n, nil
}

func (r *PGNotificationRepository) ListByRecipient(ctx context.Context, recipient model.ID, page CursorPage, _ NotificationProjection) (Page[*Notification], error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/ListByRecipient")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationRead, err)
	}

	query := "SELECT * FROM notifications WHERE recipient = $1"
	args := []any{recipient}
	if normalized.Token != nil && *normalized.Token != "" {
		id, err := DecodeCursor(*normalized.Token)
		if err != nil {
			return Page[*Notification]{}, errors.Join(ErrNotificationRead, err)
		}
		query += " AND id < $2"
		args = append(args, id)
	}
	query += fmt.Sprintf(" ORDER BY id %s LIMIT $%d", SortDirectionDesc.Cypher(), len(args)+1)
	args = append(args, normalized.FetchLimit())

	rows, err := r.db.pool.Query(ctx,
		query,
		args...,
	)
	if err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationRead, err)
	}
	defer rows.Close()

	notifications := make([]*Notification, 0, normalized.FetchLimit())

	for rows.Next() {
		n, err := scanNotificationRows(rows)
		if err != nil {
			return Page[*Notification]{}, errors.Join(ErrNotificationRead, err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return Page[*Notification]{}, errors.Join(ErrNotificationRead, err)
	}

	return PaginateSlice(notifications, normalized.Size, func(notification *Notification) model.ID {
		return notification.ID
	})
}

func (r *PGNotificationRepository) Update(ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts) (*Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Update")
	defer span.End()

	var n Notification
	row := r.db.pool.QueryRow(ctx,
		"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
		id, recipient, opts.Read,
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

func clearNotificationListByRecipient(ctx context.Context, r *redisBaseRepository, recipient model.ID) error {
	return clearNotificationsPattern(ctx, r, "ListByRecipient", recipient.String(), "*", "*", "*")
}

// RedisCachedNotificationRepository implements caching on the
// NotificationRepository.
type RedisCachedNotificationRepository struct {
	cacheRepo        *redisBaseRepository
	notificationRepo NotificationRepository
}

func (r *RedisCachedNotificationRepository) Create(ctx context.Context, opts CreateNotificationOpts) (*Notification, error) {
	if err := clearNotificationListByRecipient(ctx, r.cacheRepo, opts.Recipient); err != nil {
		return nil, err
	}

	return r.notificationRepo.Create(ctx, opts)
}

func (r *RedisCachedNotificationRepository) Get(ctx context.Context, id, recipient model.ID, proj NotificationProjection) (*Notification, error) {
	var notification *Notification
	var err error

	key := composeCacheKey(model.ResourceTypeNotification.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &notification); err != nil {
		return nil, err
	}

	if notification != nil {
		return notification, nil
	}

	if notification, err = r.notificationRepo.Get(ctx, id, recipient, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (r *RedisCachedNotificationRepository) ListByRecipient(ctx context.Context, recipient model.ID, page CursorPage, proj NotificationProjection) (Page[*Notification], error) {
	var notifications Page[*Notification]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Notification]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeNotification.String(),
		"ListByRecipient",
		recipient.String(),
		projectionCacheValue(proj),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &notifications); err != nil {
		return Page[*Notification]{}, err
	}

	if notifications.Items != nil {
		return notifications, nil
	}

	notifications, err = r.notificationRepo.ListByRecipient(ctx, recipient, normalized, proj)
	if err != nil {
		return Page[*Notification]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, notifications); err != nil {
		return Page[*Notification]{}, err
	}

	return notifications, nil
}

func (r *RedisCachedNotificationRepository) Update(ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts) (*Notification, error) {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "ListByRecipient", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return r.notificationRepo.Update(ctx, id, recipient, opts)
}

func (r *RedisCachedNotificationRepository) Delete(ctx context.Context, id, recipient model.ID) error {
	if err := clearNotificationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeNotification.String(), "ListByRecipient", "*")
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

func scanNotificationRow(row pgx.Row) (*Notification, error) {
	var n Notification
	if err := row.Scan(&n.ID, &n.Title, &n.Description, &n.Recipient, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, err
	}

	return &n, nil
}

func scanNotificationRows(rows pgx.Rows) (*Notification, error) {
	var n Notification
	if err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.Recipient, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, err
	}

	return &n, nil
}
