package repository_test

import (
	"context"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

func scanNotification(notification *repository.Notification) func(destinations ...any) error {
	return func(destinations ...any) error {
		*destinations[0].(*model.ID) = notification.ID
		*destinations[1].(*string) = notification.Title
		*destinations[2].(*string) = notification.Description
		*destinations[3].(*model.ID) = notification.Recipient
		*destinations[4].(*bool) = notification.Read
		*destinations[5].(**time.Time) = notification.CreatedAt
		*destinations[6].(**time.Time) = notification.UpdatedAt
		return nil
	}
}

func notificationScanMatchers() []any {
	return []any{
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	}
}

func TestNewNotificationRepository(t *testing.T) {
	type args struct {
		opts []repository.PGRepositoryOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "new notification repository with default options",
			args: args{
				opts: []repository.PGRepositoryOption{},
			},
		},
		{
			name: "new notification repository with no logger",
			args: args{
				opts: []repository.PGRepositoryOption{
					repository.WithPGRepositoryLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification repository with no tracer",
			args: args{
				opts: []repository.PGRepositoryOption{
					repository.WithPGRepositoryTracer(nil),
				},
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := repository.NewNotificationRepository(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func TestNotificationRepository_Create(t *testing.T) {
	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNotificationOpts) []repository.PGRepositoryOption
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateNotificationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNotificationOpts) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Create").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), opts.Title, opts.Description, opts.Recipient,
						false, gomock.Any(),
					).Return(pgconn.CommandTag{}, nil)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNotificationOpts{
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "create new notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNotificationOpts) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Create").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), opts.Title, opts.Description, opts.Recipient,
						false, gomock.Any(),
					).Return(pgconn.CommandTag{}, assert.AnError)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNotificationOpts{
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrNotificationCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			notificationRepo, err := repository.NewNotificationRepository(tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.opts)...)
			require.NoError(t, err)
			got, err := notificationRepo.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.args.opts.Title, got.Title)
				assert.Equal(t, tt.args.opts.Description, got.Description)
				assert.Equal(t, tt.args.opts.Recipient, got.Recipient)
				assert.False(t, got.Read)
				assert.NotNil(t, got.CreatedAt)
			}
		})
	}
}

func TestNotificationRepository_Get(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *repository.Notification) []repository.PGRepositoryOption
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Notification
		wantErr error
	}{
		{
			name: "get notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().
						Scan(notificationScanMatchers()...).
						DoAndReturn(scanNotification(notification))

					dbPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			want: &repository.Notification{
				ID:          notificationID,
				Title:       "test title",
				Description: "test description",
				Recipient:   recipientID,
				CreatedAt:   convert.ToPointer(time.Now()),
			},
		},
		{
			name: "get notification not found",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)

					dbPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().Scan(gomock.Any()).Return(assert.AnError)

					dbPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: repository.ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			notificationRepo, err := repository.NewNotificationRepository(tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.want)...)
			require.NoError(t, err)
			got, err := notificationRepo.Get(tt.args.ctx, tt.args.id, tt.args.recipient, repository.NotificationDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_ListByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, notifications []*repository.Notification) []repository.PGRepositoryOption
	}
	type args struct {
		ctx       context.Context
		recipient model.ID
		offset    int
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.Notification
		wantErr error
	}{
		{
			name: "get all notifications by recipient",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, notifications []*repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					rows := mockrepo.NewMockRows(ctrl)
					rows.EXPECT().Close().Return()
					rows.EXPECT().Next().Return(true).Times(limit)
					rows.EXPECT().Next().Return(false)
					rows.EXPECT().Err().Return(nil)

					for _, notification := range notifications[offset:] {
						rows.EXPECT().
							Scan(notificationScanMatchers()...).
							DoAndReturn(scanNotification(notification))
					}

					dbPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(rows, nil)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			want: []*repository.Notification{
				{
					ID:          model.MustNewID(model.ResourceTypeNotification),
					Title:       "Test",
					Description: "Test description",
					Recipient:   recipientID,
					Read:        false,
					CreatedAt:   convert.ToPointer(time.Now()),
					UpdatedAt:   nil,
				},
				{
					ID:          model.MustNewID(model.ResourceTypeNotification),
					Title:       "Test",
					Description: "Test description",
					Recipient:   recipientID,
					Read:        false,
					CreatedAt:   convert.ToPointer(time.Now()),
					UpdatedAt:   nil,
				},
			},
		},
		{
			name: "get all notifications by recipient with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, _ []*repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(mockrepo.NewMockRows(nil), assert.AnError)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			wantErr: repository.ErrNotificationRead,
		},
		{
			name: "get all notifications with scan error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, _ []*repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					rows := mockrepo.NewMockRows(ctrl)
					rows.EXPECT().Close().Return()
					rows.EXPECT().Next().Return(true).Times(1)
					rows.EXPECT().Scan(gomock.Any()).Return(assert.AnError)

					dbPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(rows, nil)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			wantErr: repository.ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			notificationRepo, err := repository.NewNotificationRepository(tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.recipient, tt.args.offset, testPageSize(tt.args.limit), tt.want)...)
			require.NoError(t, err)
			got, err := notificationRepo.ListByRecipient(tt.args.ctx, tt.args.recipient, repository.CursorPage{Size: testPageSize(tt.args.limit)}, repository.NotificationListProjection())
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got.Items)
			}
		})
	}
}

func TestNotificationRepository_Update(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts repository.UpdateNotificationOpts, notification *repository.Notification) []repository.PGRepositoryOption
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
		opts      repository.UpdateNotificationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Notification
		wantErr error
	}{
		{
			name: "update notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts repository.UpdateNotificationOpts, notification *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().
						Scan(notificationScanMatchers()...).
						DoAndReturn(scanNotification(notification))

					dbPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      repository.UpdateNotificationOpts{Read: true},
			},
			want: &repository.Notification{
				ID:          notificationID,
				Title:       "test title",
				Description: "test description",
				Recipient:   recipientID,
				Read:        true,
				CreatedAt:   convert.ToPointer(time.Now()),
				UpdatedAt:   convert.ToPointer(time.Now()),
			},
		},
		{
			name: "update notification not found",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts repository.UpdateNotificationOpts, _ *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)

					dbPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      repository.UpdateNotificationOpts{Read: true},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts repository.UpdateNotificationOpts, _ *repository.Notification) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					row := mockrepo.NewMockRow(ctrl)
					row.EXPECT().Scan(gomock.Any()).Return(assert.AnError)

					dbPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(row)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      repository.UpdateNotificationOpts{Read: true},
			},
			wantErr: repository.ErrNotificationUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			notificationRepo, err := repository.NewNotificationRepository(tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.args.opts, tt.want)...)
			require.NoError(t, err)
			got, err := notificationRepo.Update(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) []repository.PGRepositoryOption
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, nil)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
		},
		{
			name: "delete notification not found",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, pgx.ErrNoRows)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "delete notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) []repository.PGRepositoryOption {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					dbPool := mockrepo.NewMockPGPool(ctrl)
					mockDB, err := repository.NewPGDatabase(repository.WithDatabasePool(dbPool))
					require.NoError(t, err)

					dbPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, assert.AnError)

					return []repository.PGRepositoryOption{
						repository.WithPGDatabase(mockDB),
						repository.WithPGRepositoryLogger(mocklog.NewMockLogger(nil)),
						repository.WithPGRepositoryTracer(tracer),
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotificationDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			notificationRepo, err := repository.NewNotificationRepository(tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient)...)
			require.NoError(t, err)
			err = notificationRepo.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
