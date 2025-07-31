package pg

import (
	"context"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewNotificationRepository(t *testing.T) {
	type args struct {
		opts []RepositoryOption
	}
	tests := []struct {
		name    string
		args    args
		want    *NotificationRepository
		wantErr error
	}{
		{
			name: "new notification repository with default options",
			args: args{
				opts: []RepositoryOption{},
			},
			want: &NotificationRepository{
				baseRepository: &baseRepository{
					logger: log.DefaultLogger(),
					tracer: tracing.NoopTracer(),
				},
			},
		},
		{
			name: "new notification repository with no logger",
			args: args{
				opts: []RepositoryOption{
					WithRepositoryLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification repository with no tracer",
			args: args{
				opts: []RepositoryOption{
					WithRepositoryTracer(nil),
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
		baseRepository func(ctx context.Context, ctrl *gomock.Controller, notification *model.Notification) *baseRepository
	}
	type args struct {
		ctx          context.Context
		notification *model.Notification
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, notification *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), notification.Title, notification.Description, notification.Recipient.String(),
						notification.Read, gomock.Any(),
					).Return(pgconn.CommandTag{}, nil)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				notification: &model.Notification{
					ID:          model.MustNewNilID(model.ResourceTypeNotification),
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
		},
		{
			name: "create new notification with error",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, notification *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"INSERT INTO notifications (id, title, description, recipient, read, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
						gomock.Any(), notification.Title, notification.Description, notification.Recipient.String(),
						notification.Read, gomock.Any(),
					).Return(pgconn.CommandTag{}, assert.AnError)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				notification: &model.Notification{
					ID:          model.MustNewNilID(model.ResourceTypeNotification),
					Title:       "test notification",
					Description: "test description",
					Recipient:   model.MustNewNilID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrNotificationCreate,
		},
		{
			name: "create new invalid notification",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				notification: &model.Notification{
					ID:          model.MustNewNilID(model.ResourceTypeNotification),
					Title:       "",
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
			defer ctrl.Finish()
			notificationRepo := &NotificationRepository{
				baseRepository: tt.fields.baseRepository(tt.args.ctx, ctrl, tt.args.notification),
			}
			err := notificationRepo.Create(tt.args.ctx, tt.args.notification)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNotificationRepository_Get(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeNotification)

	type fields struct {
		baseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *model.Notification) *baseRepository
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
		want    *model.Notification
		wantErr error
	}{
		{
			name: "get notification",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, notification *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(dest ...any) error {
							dest[0].(*pgID).ID = notification.ID
							*(dest[1].(*string)) = notification.Title
							*(dest[2].(*string)) = notification.Description
							dest[3].(*pgID).ID = notification.Recipient
							*(dest[4].(*bool)) = notification.Read
							*(dest[5].(**time.Time)) = notification.CreatedAt
							*(dest[6].(**time.Time)) = notification.UpdatedAt
							return nil
						})

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						[]any{id.String(), recipient.String()},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			want: &model.Notification{
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(pgx.ErrNoRows)

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						[]any{id.String(), recipient.String()},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().QueryRow(ctx,
						"SELECT * FROM notifications WHERE id = $1 AND recipient = $2",
						[]any{id.String(), recipient.String()},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
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
		{
			name: "get notification with invalid notification",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: repository.ErrNotificationRead,
		},
		{
			name: "get notification with invalid recipient",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: repository.ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &NotificationRepository{
				baseRepository: tt.fields.baseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.want),
			}
			got, err := notificationRepo.Get(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_GetAllByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeNotification)

	type fields struct {
		baseRepository func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, notifications []*model.Notification) *baseRepository
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
		want    []*model.Notification
		wantErr error
	}{
		{
			name: "get all notifications",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, notifications []*model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRows := mock.NewMockRows(ctrl)
					mockRows.EXPECT().Close().Return()
					mockRows.EXPECT().Next().Return(true).Times(limit)
					mockRows.EXPECT().Next().Return(false)

					for _, notification := range notifications[offset:] {
						mockRows.EXPECT().
							Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							DoAndReturn(func(dest ...any) error {
								dest[0].(*pgID).ID = notification.ID
								*(dest[1].(*string)) = notification.Title
								*(dest[2].(*string)) = notification.Description
								dest[3].(*pgID).ID = notification.Recipient
								*(dest[4].(*bool)) = notification.Read
								*(dest[5].(**time.Time)) = notification.CreatedAt
								*(dest[6].(**time.Time)) = notification.UpdatedAt
								return nil
							}).
							Times(1)
					}

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 LIMIT $2 OFFSET $3",
						[]any{recipient.String(), limit, offset},
					).Return(mockRows, nil)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
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
			want: []*model.Notification{
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
			name: "get all notifications with error",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, _ []*model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 LIMIT $2 OFFSET $3",
						[]any{recipient.String(), limit, offset},
					).Return(mock.NewMockRows(nil), assert.AnError)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
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
			wantErr: repository.ErrNotificationRead,
		},
		{
			name: "get all notifications with invalid ID",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDB, err := NewDatabase(WithDatabasePool(mock.NewMockPool(nil)))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				recipient: model.ID{},
				limit:     2,
				offset:    0,
			},
			wantErr: repository.ErrNotificationRead,
		},
		{
			name: "get all notifications with scan error",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, recipient model.ID, offset, limit int, _ []*model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRows := mock.NewMockRows(ctrl)
					mockRows.EXPECT().Close().Return()
					mockRows.EXPECT().Next().Return(true).Times(1)
					mockRows.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().Query(ctx,
						"SELECT * FROM notifications WHERE recipient = $1 LIMIT $2 OFFSET $3",
						[]any{recipient.String(), limit, offset},
					).Return(mockRows, nil)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
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
			wantErr: repository.ErrNotificationRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &NotificationRepository{
				baseRepository: tt.fields.baseRepository(tt.args.ctx, ctrl, tt.args.recipient, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := notificationRepo.GetAllByRecipient(tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_Update(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeNotification)

	type fields struct {
		baseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, read bool, notification *model.Notification) *baseRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
		read      bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Notification
		wantErr error
	}{
		{
			name: "update notification",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, read bool, notification *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(dest ...any) error {
							dest[0].(*pgID).ID = notification.ID
							*(dest[1].(*string)) = notification.Title
							*(dest[2].(*string)) = notification.Description
							dest[3].(*pgID).ID = notification.Recipient
							*(dest[4].(*bool)) = notification.Read
							*(dest[5].(**time.Time)) = notification.CreatedAt
							*(dest[6].(**time.Time)) = notification.UpdatedAt
							return nil
						})

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						[]any{id.String(), recipient.String(), read},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			want: &model.Notification{
				ID:          notificationID,
				Title:       "test title",
				Description: "test description",
				Recipient:   recipientID,
				CreatedAt:   convert.ToPointer(time.Now()),
			},
		},
		{
			name: "update notification not found",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, read bool, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(pgx.ErrNoRows)

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						[]any{id.String(), recipient.String(), read},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
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
			name: "update notification with error",
			fields: fields{
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID, read bool, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockRow := mock.NewMockRow(ctrl)
					mockRow.EXPECT().
						Scan(gomock.Any()).
						Return(assert.AnError)

					mockDBPool.EXPECT().QueryRow(ctx,
						"UPDATE notifications SET read = $3, updated_at = timezone('utc', now()) WHERE id = $1 AND recipient = $2 RETURNING *",
						[]any{id.String(), recipient.String(), read},
					).Return(mockRow)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: repository.ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid notification ID",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID, _ bool, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: repository.ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid recipient ID",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID, _ bool, _ *model.Notification) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: repository.ErrNotificationUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &NotificationRepository{
				baseRepository: tt.fields.baseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient, tt.args.read, tt.want),
			}
			got, err := notificationRepo.Update(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.read)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	type fields struct {
		baseRepository func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *baseRepository
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id.String(), recipient.String(),
					).Return(pgconn.CommandTag{}, nil)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id.String(), recipient.String(),
					).Return(pgconn.CommandTag{}, pgx.ErrNoRows)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
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
				baseRepository: func(ctx context.Context, ctrl *gomock.Controller, id, recipient model.ID) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(ctrl)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					mockDBPool.EXPECT().Exec(ctx,
						"DELETE FROM notifications WHERE id = $1 AND recipient = $2",
						id.String(), recipient.String(),
					).Return(pgconn.CommandTag{}, assert.AnError)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
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
		{
			name: "delete notification with invalid notification ID",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.ID{},
				recipient: model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotificationDelete,
		},
		{
			name: "delete notification with invalid recipient ID",
			fields: fields{
				baseRepository: func(ctx context.Context, _ *gomock.Controller, _, _ model.ID) *baseRepository {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.pg.NotificationRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					mockDBPool := mock.NewMockPool(nil)
					mockDB, err := NewDatabase(WithDatabasePool(mockDBPool))
					require.NoError(t, err)

					return &baseRepository{
						db:     mockDB,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewNilID(model.ResourceTypeNotification),
				recipient: model.ID{},
			},
			wantErr: repository.ErrNotificationDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			notificationRepo := &NotificationRepository{
				baseRepository: tt.fields.baseRepository(tt.args.ctx, ctrl, tt.args.id, tt.args.recipient),
			}
			err := notificationRepo.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
