package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateNotificationOpts(recipient model.ID) CreateNotificationOpts {
	return CreateNotificationOpts{
		Title:       "test notification",
		Description: "test description",
		Recipient:   recipient,
	}
}

func TestCreateNotificationOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    CreateNotificationOpts
		wantErr error
	}{
		{
			name: "valid notification",
			opts: CreateNotificationOpts{
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
		},
		{
			name: "invalid notification title",
			opts: CreateNotificationOpts{
				Title:       "he",
				Description: "Test description",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: model.ErrInvalidNotificationDetails,
		},
		{
			name: "invalid notification description",
			opts: CreateNotificationOpts{
				Title:       "Test Notification",
				Description: "Test",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: model.ErrInvalidNotificationDetails,
		},
		{
			name: "invalid recipient ID",
			opts: CreateNotificationOpts{
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   model.ID{},
			},
			wantErr: model.ErrInvalidNotificationRecipient,
		},
		{
			name: "invalid recipient ID type",
			opts: CreateNotificationOpts{
				Title:       "Test Notification",
				Description: "Test description",
				Recipient:   model.MustNewNilID(model.ResourceTypeOrganization),
			},
			wantErr: model.ErrInvalidNotificationRecipient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, tt.opts.Validate(), tt.wantErr)
		})
	}
}

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
				repo: repository.NewMockNotificationRepository(nil),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &notificationService{
				baseService: &baseService{
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				notificationRepo: repository.NewMockNotificationRepository(nil),
			},
		},
		{
			name: "new notification service with invalid options",
			args: args{
				repo: repository.NewMockNotificationRepository(nil),
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification service with no notification repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
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
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, opts CreateNotificationOpts) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateNotificationOpts) repository.NotificationRepository
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
			name: "create notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNotificationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateNotificationOpts) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Create(ctx, repository.CreateNotificationOpts{
						Title:       opts.Title,
						Description: opts.Description,
						Recipient:   opts.Recipient,
					}).Return(testModel.NewRepositoryNotification(opts.Recipient), nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				opts: newCreateNotificationOpts(recipientID),
			},
		},
		{
			name: "create notification with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNotificationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateNotificationOpts) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Create(ctx, repository.CreateNotificationOpts{
						Title:       opts.Title,
						Description: opts.Description,
						Recipient:   opts.Recipient,
					}).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				opts: newCreateNotificationOpts(recipientID),
			},
			wantErr: ErrNotificationCreate,
		},
		{
			name: "create notification with invalid notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNotificationOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateNotificationOpts) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNotificationOpts{
					Recipient: model.ID{},
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
			s := &notificationService{
				baseService:      tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := s.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNotificationService_Get(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *repository.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *repository.Notification) repository.NotificationRepository
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
		repoN   *repository.Notification
		want    *Notification
		wantErr error
	}{
		{
			name: "get notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, recipient).Return(notification, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			repoN: &repository.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
			},
			want: &Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
			},
		},
		{
			name: "get notification with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, _ *repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
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
			name: "get notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
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
			name: "get notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: ErrNotificationGet,
		},
		{
			name: "get notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &notificationService{
				baseService:      tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.repoN),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.repoN),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_GetAllByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, notifications []*repository.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, notifications []*repository.Notification) repository.NotificationRepository
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
		repoN   []*repository.Notification
		want    []*Notification
		wantErr error
	}{
		{
			name: "get all notifications by recipient",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/GetAllByRecipient", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, notifications []*repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().GetAllByRecipient(ctx, recipient, offset, limit).Return(notifications, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    0,
				limit:     10,
			},
			repoN: []*repository.Notification{
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
			want: []*Notification{
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
			name: "get all notifications by recipient with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/GetAllByRecipient", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, offset, limit int, _ []*repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().GetAllByRecipient(ctx, recipient, offset, limit).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    0,
				limit:     10,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
		{
			name: "get all notifications by recipient with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/GetAllByRecipient", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				recipient: model.ID{},
				offset:    0,
				limit:     10,
			},
			wantErr: ErrNotificationGetAllByRecipient,
		},
		{
			name: "get all notifications by recipient with invalid pagination",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/GetAllByRecipient", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				offset:    -1,
				limit:     0,
			},
			wantErr: ErrInvalidPaginationParams,
		},
		{
			name: "get all notifications by recipient with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _, _ int, _ []*repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/GetAllByRecipient", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				recipient: recipientID,
				offset:    0,
				limit:     10,
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Align want IDs with repo fixtures when both are set with matching length.
			if len(tt.repoN) > 0 && len(tt.want) == len(tt.repoN) {
				for i := range tt.want {
					tt.want[i].ID = tt.repoN[i].ID
				}
			}

			s := &notificationService{
				baseService:      tt.fields.baseService(ctrl, tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit, tt.repoN),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit, tt.repoN),
			}
			got, err := s.GetAllByRecipient(tt.args.ctx, tt.args.recipient, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_Update(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts, notification *repository.Notification) *baseService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts, notification *repository.Notification) repository.NotificationRepository
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
		repoN   *repository.Notification
		want    *Notification
		wantErr error
	}{
		{
			name: "update notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts, notification *repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, repository.UpdateNotificationOpts{Read: opts.Read}).Return(notification, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			repoN: &repository.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
				Read:        true,
			},
			want: &Notification{
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, repository.UpdateNotificationOpts{Read: opts.Read}).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNotificationUpdate,
		},
		{
			name: "update notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
				opts:      UpdateNotificationOpts{Read: true},
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &notificationService{
				baseService:      tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts, tt.repoN),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts, tt.repoN),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_Delete(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := repository.NewMockNotificationRepository(ctrl)
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
			name: "delete notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
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
			name: "delete notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: ErrNotificationDelete,
		},
		{
			name: "delete notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return repository.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &notificationService{
				baseService:      tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient),
				notificationRepo: tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
