package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewNotificationRepository(t *testing.T) {
	type args struct {
		opts []PGRepositoryOption
	}
	tests := []struct {
		name    string
		args    args
		want    *PGNotificationRepository
		wantErr error
	}{
		{
			name: "new notification repository with default options",
			args: args{
				opts: []PGRepositoryOption{},
			},
			want: &PGNotificationRepository{
				pgBaseRepository: &pgBaseRepository{
					logger: log.DefaultLogger(),
					tracer: tracing.NoopTracer(),
				},
			},
		},
		{
			name: "new notification repository with no logger",
			args: args{
				opts: []PGRepositoryOption{
					WithPGRepositoryLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification repository with no tracer",
			args: args{
				opts: []PGRepositoryOption{
					WithPGRepositoryTracer(nil),
				},
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNotificationRepository(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_Create(t *testing.T) {
	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, opts CreateNotificationOpts) *pgBaseRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateNotificationOpts
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, opts CreateNotificationOpts) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Create").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), opts.Title, opts.Description, opts.Recipient,
						false, gomock.Any(),
					).Return(pgconn.CommandTag{}, nil)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNotificationOpts{
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "create new notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, opts CreateNotificationOpts) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Create").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), opts.Title, opts.Description, opts.Recipient,
						false, gomock.Any(),
					).Return(pgconn.CommandTag{}, assert.AnError)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNotificationOpts{
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
			wantErr: ErrNotificationCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &PGNotificationRepository{
				pgBaseRepository: tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.opts),
			}
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
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *Notification) *pgBaseRepository
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
		want    *Notification
		wantErr error
	}{
		{
			name: "get notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(dest ...any) error {
							*(dest[0].(*model.ID)) = notification.ID
							*(dest[1].(*string)) = notification.Title
							*(dest[2].(*string)) = notification.Description
							*(dest[3].(*model.ID)) = notification.Recipient
							*(dest[4].(*bool)) = notification.Read
							*(dest[5].(**time.Time)) = notification.CreatedAt
							*(dest[6].(**time.Time)) = notification.UpdatedAt
							return nil
						})

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			want: &Notification{
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(pgx.ErrNoRows)

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Get").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &PGNotificationRepository{
				pgBaseRepository: tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.want),
			}
			got, err := notificationRepo.Get(tt.args.ctx, tt.args.id, tt.args.recipient, NotificationDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_ListByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, notifications []*Notification) *pgBaseRepository
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
		want    []*Notification
		wantErr error
	}{
		{
			name: "get all notifications by recipient",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, notifications []*Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRows := mock.NewPGRows(ctrl)
					mockRows.EXPECT().Close().Return()
					mockRows.EXPECT().Next().Return(true).Times(limit)
					mockRows.EXPECT().Next().Return(false)
					mockRows.EXPECT().Err().Return(nil)

					for _, notification := range notifications[offset:] {
						mockRows.EXPECT().
							Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							DoAndReturn(func(dest ...any) error {
								*(dest[0].(*model.ID)) = notification.ID
								*(dest[1].(*string)) = notification.Title
								*(dest[2].(*string)) = notification.Description
								*(dest[3].(*model.ID)) = notification.Recipient
								*(dest[4].(*bool)) = notification.Read
								*(dest[5].(**time.Time)) = notification.CreatedAt
								*(dest[6].(**time.Time)) = notification.UpdatedAt
								return nil
							}).
							Times(1)
					}

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(mockRows, nil)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			want: []*Notification{
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, _ []*Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(mock.NewPGRows(nil), assert.AnError)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			wantErr: ErrNotificationRead,
		},
		{
			name: "get all notifications with scan error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, _, limit int, _ []*Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/ListByRecipient").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRows := mock.NewPGRows(ctrl)
					mockRows.EXPECT().Close().Return()
					mockRows.EXPECT().Next().Return(true).Times(1)
					mockRows.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 ORDER BY id DESC LIMIT $2",
						recipient, limit+1,
					).Return(mockRows, nil)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: recipientID,
				limit:     2,
				offset:    0,
			},
			wantErr: ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &PGNotificationRepository{
				pgBaseRepository: tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.recipient, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := notificationRepo.ListByRecipient(tt.args.ctx, tt.args.recipient, CursorPage{Size: testPageSize(tt.args.limit)}, NotificationListProjection())
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
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts UpdateNotificationOpts, notification *Notification) *pgBaseRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
		opts      UpdateNotificationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Notification
		wantErr error
	}{
		{
			name: "update notification",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts UpdateNotificationOpts, notification *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(dest ...any) error {
							*(dest[0].(*model.ID)) = notification.ID
							*(dest[1].(*string)) = notification.Title
							*(dest[2].(*string)) = notification.Description
							*(dest[3].(*model.ID)) = notification.Recipient
							*(dest[4].(*bool)) = notification.Read
							*(dest[5].(**time.Time)) = notification.CreatedAt
							*(dest[6].(**time.Time)) = notification.UpdatedAt
							return nil
						})

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			want: &Notification{
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts UpdateNotificationOpts, _ *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(pgx.ErrNoRows)

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, opts UpdateNotificationOpts, _ *Notification) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Update").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewPGRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						id, recipient, opts.Read,
					).Return(mockRow)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNotificationUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &PGNotificationRepository{
				pgBaseRepository: tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.args.opts, tt.want),
			}
			got, err := notificationRepo.Update(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	type fields struct {
		pgBaseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *pgBaseRepository
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, nil)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
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
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, pgx.ErrNoRows)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "delete notification with error",
			fields: fields{
				pgBaseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *pgBaseRepository {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.pg.NotificationRepository/Delete").Return(ctx, span)

					mockDBPool := mock.NewPGPool(ctrl)
					mockDB, err := NewPGDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id, recipient,
					).Return(pgconn.CommandTag{}, assert.AnError)

					return &pgBaseRepository{
						db:     mockDB,
						logger: mock.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: ErrNotificationDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &PGNotificationRepository{
				pgBaseRepository: tt.fields.pgBaseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient),
			}
			err := notificationRepo.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
