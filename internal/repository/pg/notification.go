package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// NotificationRepository is a repository for managing notifications.
type NotificationRepository struct {
	*baseRepository
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Create")
	defer span.End()

	if err := notification.Validate(); err != nil {
		return errors.Join(repository.ErrNotificationCreate, err)
	}

	createdAt := time.Now().UTC().Round(time.Microsecond)

	notification.ID = model.MustNewID(model.ResourceTypeNotification)
	notification.Read = false
	notification.CreatedAt = &createdAt
	notification.UpdatedAt = nil

	_, err := r.db.pool.Exec(ctx,
		"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		notification.ID.String(), notification.Title, notification.Description, notification.Recipient.String(),
		notification.Read, createdAt,
	)

	if err != nil {
		return errors.Join(repository.ErrNotificationCreate, err)
	}

	return nil
}

func (r *NotificationRepository) Get(ctx context.Context, id, recipient model.ID) (*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(repository.ErrNotificationRead, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(repository.ErrNotificationRead, err)
	}

	var nid, rid pgID
	var n model.Notification
	row := r.db.pool.QueryRow(ctx, "SELECT * FROM notifications WHERE id = $1 AND recipient = $2", id.String(), recipient.String())
	if err := row.Scan(&nid, &n.Title, &n.Description, &rid, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, errors.Join(repository.ErrNotificationRead, err)
	}

	n.ID = nid.ID
	n.Recipient = rid.ID
	return &n, nil
}

func (r *NotificationRepository) GetAllByRecipient(ctx context.Context, recipient model.ID, offset, limit int) ([]*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/GetAllByRecipient")
	defer span.End()

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(repository.ErrNotificationRead, err)
	}

	rows, err := r.db.pool.Query(ctx,
		"SELECT * FROM notifications WHERE recipient = $1 LIMIT $2 OFFSET $3",
		recipient.String(), limit, offset,
	)
	if err != nil {
		return nil, errors.Join(repository.ErrNotificationRead, err)
	}
	defer rows.Close()

	notifications := make([]*model.Notification, 0)

	for rows.Next() {
		var nid, rid pgID

		var n model.Notification
		if err := rows.Scan(&nid, &n.Title, &n.Description, &rid, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, errors.Join(repository.ErrNotificationRead, err)
		}

		n.ID = nid.ID
		n.Recipient = rid.ID
		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (r *NotificationRepository) Update(ctx context.Context, id, recipient model.ID, read bool) (*model.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Update")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(repository.ErrNotificationUpdate, err)
	}

	if err := recipient.Validate(); err != nil {
		return nil, errors.Join(repository.ErrNotificationUpdate, err)
	}

	var nid, rid pgID
	var n model.Notification
	row := r.db.pool.QueryRow(ctx,
		"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
		id.String(), recipient.String(), read,
	)
	if err := row.Scan(&nid, &n.Title, &n.Description, &rid, &n.Read, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, errors.Join(repository.ErrNotificationUpdate, err)
	}

	n.ID = nid.ID
	n.Recipient = rid.ID
	return &n, nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id, recipient model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.NotificationRepository/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(repository.ErrNotificationDelete, err)
	}

	if err := recipient.Validate(); err != nil {
		return errors.Join(repository.ErrNotificationDelete, err)
	}

	_, err := r.db.pool.Exec(ctx,
		"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
		id.String(), recipient.String(),
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ErrNotFound
		}
		return errors.Join(repository.ErrNotificationDelete, err)
	}

	return nil
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(opts ...RepositoryOption) (*NotificationRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &NotificationRepository{
		baseRepository: baseRepo,
	}, nil
}
