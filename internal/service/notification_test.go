package service

import (
	"context"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNewNotificationService(t *testing.T) {
	type args struct {
		repo repository.NotificationRepository
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    NotificationService
		wantErr error
	}{
		{
			name: "new notification service",
			args: args{
				repo: mock.NewNotificationRepository(nil),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &notificationService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				notificationRepo: mock.NewNotificationRepository(nil),
			},
		},
		{
			name: "new notification service with invalid options",
			args: args{
				repo: mock.NewNotificationRepository(nil),
				opts: []Option{
					WithLogger(nil),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification service with no notification repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoNotificationRepository,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNotificationService(tt.args.repo, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_Create(t *testing.T) {
	type fields struct {
		baseService      func(ctx context.Context, notification *model.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, notification *model.Notification) repository.NotificationRepository
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
			name: "create notification",
			fields: fields{
				baseService: func(ctx context.Context, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, notification *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Create(ctx, notification).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				notification: testModel.NewNotification(model.MustNewID(model.ResourceTypeUser)),
			},
		},
		{
			name: "create notification with error",
			fields: fields{
				baseService: func(ctx context.Context, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, notification *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Create(ctx, notification).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				notification: testModel.NewNotification(model.MustNewID(model.ResourceTypeUser)),
			},
			wantErr: ErrNotificationCreate,
		},
		{
			name: "create notification with invalid notification",
			fields: fields{
				baseService: func(ctx context.Context, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				notification: &model.Notification{Recipient: model.ID{}},
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
			s := &notificationService{
				baseService:      tt.fields.baseService(tt.args.ctx, tt.args.notification),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.notification),
			}
			err := s.Create(tt.args.ctx, tt.args.notification)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNotificationService_Get(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctx context.Context, id, recipient model.ID, notification *model.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *model.Notification) repository.NotificationRepository
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
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, recipient).Return(notification, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			want: &model.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
			},
		},
		{
			name: "get notification with error",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, _ *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, recipient).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotificationGet,
		},
		{
			name: "get notification for other user",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotificationGet,
		},
		{
			name: "get notification with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: ErrNotificationGet,
		},
		{
			name: "get notification with invalid recipient",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: ErrNotificationGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &notificationService{
				baseService:      tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.recipient, tt.want),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.want),
			}
			notification, err := s.Get(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, notification)
		})
	}
}

func TestNotificationService_GetAllByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctx context.Context, recipient model.ID, offset, limit int, notifications []*model.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, notifications []*model.Notification) repository.NotificationRepository
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
			name: "get notifications",
			fields: fields{
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, notifications []*model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().GetAllByRecipient(ctx, recipient, offset, limit).Return(notifications, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    0,
				limit:     2,
			},
			want: []*model.Notification{
				{
					ID:          model.MustNewID(model.ResourceTypeNotification),
					Title:       "test",
					Description: "test notification",
					Recipient:   recipientID,
				},
				{
					ID:          model.MustNewID(model.ResourceTypeNotification),
					Title:       "test",
					Description: "test notification",
					Recipient:   recipientID,
				},
			},
		},
		{
			name: "get notifications with error",
			fields: fields{
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, _ []*model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().GetAllByRecipient(ctx, recipient, offset, limit).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    0,
				limit:     2,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
		{
			name: "get notifications for other user",
			fields: fields{
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				recipient: recipientID,
				offset:    0,
				limit:     2,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
		{
			name: "get notifications with invalid recipient",
			fields: fields{
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				recipient: model.ID{},
				offset:    0,
				limit:     2,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
		{
			name: "get notifications with invalid pagination params",
			fields: fields{
				baseService: func(ctx context.Context, _ model.ID, _, _ int, _ []*model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/GetAllByRecipient", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    0,
				limit:     0,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &notificationService{
				baseService:      tt.fields.baseService(tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit, tt.want),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit, tt.want),
			}
			notification, err := s.GetAllByRecipient(tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, notification)
		})
	}
}

func TestNotificationService_Update(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctx context.Context, id, recipient model.ID, read bool, notification *model.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, read bool, notification *model.Notification) repository.NotificationRepository
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
				baseService: func(ctx context.Context, _, _ model.ID, _ bool, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, read bool, notification *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, read).Return(notification, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				read:      true,
			},
			want: &model.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
				Read:        true,
			},
		},
		{
			name: "update notification with error",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ bool, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, read bool, _ *model.Notification) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, read).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				read:      true,
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification for other user",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ bool, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
				read:      true,
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ bool, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
				read:      true,
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid recipient",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID, _ bool, _ *model.Notification) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool, _ *model.Notification) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
				read:      true,
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
			s := &notificationService{
				baseService:      tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.read, tt.want),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.read, tt.want),
			}
			notification, err := s.Update(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.read)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, notification)
		})
	}
}

func TestNotificationService_Delete(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctx context.Context, id, recipient model.ID) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository
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
				baseService: func(ctx context.Context, _, _ model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id, recipient).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
		},
		{
			name: "delete notification with error",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := mock.NewNotificationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id, recipient).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotificationDelete,
		},
		{
			name: "delete notification for other user",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNotificationDelete,
		},
		{
			name: "delete notification with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: ErrNotificationDelete,
		},
		{
			name: "delete notification with invalid recipient",
			fields: fields{
				baseService: func(ctx context.Context, _, _ model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.notificationService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mock.NewNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
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
			s := &notificationService{
				baseService:      tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.recipient),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
