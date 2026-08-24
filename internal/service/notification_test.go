package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateNotificationOpts(recipient model.ID) service.CreateNotificationOpts {
	return service.CreateNotificationOpts{
		Title:       "test notification",
		Description: "test description",
		Recipient:   recipient,
	}
}

func TestCreateNotificationOpts_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    service.CreateNotificationOpts
		wantErr error
	}{
		{
			name: "valid notification",
			opts: service.CreateNotificationOpts{
				Title:       "Test service.Notification",
				Description: "Test description",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
		},
		{
			name: "invalid notification title",
			opts: service.CreateNotificationOpts{
				Title:       "he",
				Description: "Test description",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: model.ErrInvalidNotificationDetails,
		},
		{
			name: "invalid notification description",
			opts: service.CreateNotificationOpts{
				Title:       "Test service.Notification",
				Description: "Test",
				Recipient:   model.MustNewNilID(model.ResourceTypeUser),
			},
			wantErr: model.ErrInvalidNotificationDetails,
		},
		{
			name: "invalid recipient ID",
			opts: service.CreateNotificationOpts{
				Title:       "Test service.Notification",
				Description: "Test description",
				Recipient:   model.ID{},
			},
			wantErr: model.ErrInvalidNotificationRecipient,
		},
		{
			name: "invalid recipient ID type",
			opts: service.CreateNotificationOpts{
				Title:       "Test service.Notification",
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
	tests := []struct {
		name    string
		build   func() (service.NotificationService, error)
		wantErr error
	}{
		{
			name: "new notification service",
			build: func() (service.NotificationService, error) {
				return service.NewNotificationService(
					mockrepo.NewMockNotificationRepository(nil),
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				)
			},
		},
		{
			name: "new notification service with invalid options",
			build: func() (service.NotificationService, error) {
				return service.NewNotificationService(
					mockrepo.NewMockNotificationRepository(nil),
					service.WithLogger(nil),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				)
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new notification service with no notification repository",
			build: func() (service.NotificationService, error) {
				return service.NewNotificationService(
					nil,
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				)
			},
			wantErr: service.ErrNoNotificationRepository,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.build()
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestNotificationService_Create(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateNotificationOpts) service.NotificationService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateNotificationOpts) repository.NotificationRepository
	}
	type args struct {
		ctx  context.Context
		opts service.CreateNotificationOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateNotificationOpts) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateNotificationOpts) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateNotificationOpts) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateNotificationOpts) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
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
			wantErr: service.ErrNotificationCreate,
		},
		{
			name: "create notification with invalid notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateNotificationOpts) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Create", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ service.CreateNotificationOpts) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: service.CreateNotificationOpts{
					Recipient: model.ID{},
				},
			},
			wantErr: service.ErrNotificationCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts)
			service.SetNotificationServiceRepo(s, tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.opts))
			_, err := s.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNotificationService_Get(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *repository.Notification) service.NotificationService
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
		want    *service.Notification
		wantErr error
	}{
		{
			name: "get notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, notification *repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, recipient, repository.NotificationDetailProjection()).Return(notification, nil)
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
			want: &service.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
			},
		},
		{
			name: "get notification with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, _ *repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, recipient, repository.NotificationDetailProjection()).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: service.ErrNotificationGet,
		},
		{
			name: "get notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: service.ErrNotificationGet,
		},
		{
			name: "get notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: service.ErrNotificationGet,
		},
		{
			name: "get notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.repoN)
			service.SetNotificationServiceRepo(s, tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.repoN))
			got, err := s.Get(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_ListByRecipient(t *testing.T) {
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, page service.CursorPage, notifications []*repository.Notification) service.NotificationService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, page service.CursorPage, notifications []*repository.Notification) repository.NotificationRepository
	}
	type args struct {
		ctx       context.Context
		recipient model.ID
		page      service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		repoN   []*repository.Notification
		want    service.Page[*service.Notification]
		wantErr error
	}{
		{
			name: "list notifications by recipient",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/ListByRecipient", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, page service.CursorPage, notifications []*repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().ListByRecipient(ctx, recipient, page, repository.NotificationListProjection()).Return(repository.Page[*repository.Notification]{Items: notifications}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				page:      service.CursorPage{Size: 10},
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
			want: service.Page[*service.Notification]{Items: []*service.Notification{
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
			}},
		},
		{
			name: "list notifications by recipient with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/ListByRecipient", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, recipient model.ID, page service.CursorPage, _ []*repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().ListByRecipient(ctx, recipient, page, repository.NotificationListProjection()).Return(repository.Page[*repository.Notification]{}, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNotificationListByRecipient,
		},
		{
			name: "list notifications by recipient with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/ListByRecipient", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				recipient: model.ID{},
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNotificationListByRecipient,
		},
		{
			name: "list notifications by recipient with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/ListByRecipient", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				recipient: recipientID,
				page:      service.CursorPage{Size: -1},
			},
			wantErr: service.ErrNotificationListByRecipient,
		},
		{
			name: "list notifications by recipient with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/ListByRecipient", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ service.CursorPage, _ []*repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				recipient: recipientID,
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Align want IDs with repo fixtures when both are set with matching length.
			if len(tt.repoN) > 0 && len(tt.want.Items) == len(tt.repoN) {
				for i := range tt.want.Items {
					tt.want.Items[i].ID = tt.repoN[i].ID
				}
			}

			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.recipient, tt.args.page, tt.repoN)
			service.SetNotificationServiceRepo(s, tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.recipient, tt.args.page, tt.repoN))
			got, err := s.ListByRecipient(tt.args.ctx, tt.args.recipient, tt.args.page)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotificationService_Update(t *testing.T) {
	notificationID := model.MustNewID(model.ResourceTypeNotification)
	recipientID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts service.UpdateNotificationOpts, notification *repository.Notification) service.NotificationService
		notificationRepo func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts service.UpdateNotificationOpts, notification *repository.Notification) repository.NotificationRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		recipient model.ID
		opts      service.UpdateNotificationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		repoN   *repository.Notification
		want    *service.Notification
		wantErr error
	}{
		{
			name: "update notification",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts service.UpdateNotificationOpts, notification *repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, repository.UpdateNotificationOpts{Read: opts.Read}).Return(notification, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				opts:      service.UpdateNotificationOpts{Read: true},
			},
			repoN: &repository.Notification{
				ID:          notificationID,
				Title:       "test",
				Description: "test notification",
				Recipient:   recipientID,
				Read:        true,
			},
			want: &service.Notification{
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID, opts service.UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, recipient, repository.UpdateNotificationOpts{Read: opts.Read}).Return(nil, assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
				opts:      service.UpdateNotificationOpts{Read: true},
			},
			wantErr: service.ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
				opts:      service.UpdateNotificationOpts{Read: true},
			},
			wantErr: service.ErrNotificationUpdate,
		},
		{
			name: "update notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
				opts:      service.UpdateNotificationOpts{Read: true},
			},
			wantErr: service.ErrNotificationUpdate,
		},
		{
			name: "update notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ service.UpdateNotificationOpts, _ *repository.Notification) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
				opts:      service.UpdateNotificationOpts{Read: true},
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts, tt.repoN)
			service.SetNotificationServiceRepo(s, tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient, tt.args.opts, tt.repoN))
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
		baseService      func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) service.NotificationService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, recipient model.ID) repository.NotificationRepository {
					repo := mockrepo.NewMockNotificationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id, recipient).Return(assert.AnError)
					return repo
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: service.ErrNotificationDelete,
		},
		{
			name: "delete notification with invalid notification ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, recipientID),
				id:        model.ID{},
				recipient: recipientID,
			},
			wantErr: service.ErrNotificationDelete,
		},
		{
			name: "delete notification with invalid recipient ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.ID{}),
				id:        notificationID,
				recipient: model.ID{},
			},
			wantErr: service.ErrNotificationDelete,
		},
		{
			name: "delete notification with permission denied",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ model.ID) service.NotificationService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.notificationService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.NotificationService {
						svc, err := service.NewNotificationService(
							mockrepo.NewMockNotificationRepository(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
				notificationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.NotificationRepository {
					return mockrepo.NewMockNotificationRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				id:        notificationID,
				recipient: recipientID,
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient)
			service.SetNotificationServiceRepo(s, tt.fields.notificationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.recipient))
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.recipient)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
