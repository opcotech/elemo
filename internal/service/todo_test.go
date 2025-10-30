package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNewTodoService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    TodoService
		wantErr error
	}{
		{
			name: "new todo service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTodoRepository(mock.NewTodoRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &todoService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					todoRepo:          mock.NewTodoRepository(nil),
					permissionService: mock.NewPermissionService(nil),
					licenseService:    mock.NewMockLicenseService(nil),
				},
			},
		},
		{
			name: "new todo service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTodoRepository(mock.NewTodoRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new todo service with no todo repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoTodoRepository,
		},
		{
			name: "new todo service with no permission repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTodoRepository(mock.NewTodoRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new todo service with no license service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTodoRepository(mock.NewTodoRepository(nil)),
					WithPermissionService(mock.NewPermissionService(nil)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewTodoService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTodoService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeTodo)
	peerID := model.MustNewID(model.ResourceTypeTodo)

	type args struct {
		ctx  context.Context
		todo *model.Todo
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr error
	}{
		{
			name: "create todo",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, todo).Return(nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
		},
		{
			name: "create todo for peer",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().HasAnyRelation(ctx, peerID, userID).Return(true, nil)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, todo).Return(nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
		},
		{
			name: "create todo with invalid todo",
			args: args{
				ctx:  context.Background(),
				todo: &model.Todo{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mock.NewTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoCreate,
		},
		{
			name: "create todo with expired license",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					todoRepo := mock.NewTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create todo with license service error",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					todoRepo := mock.NewTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create todo",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, todo).Return(assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoCreate,
		},
		{
			name: "create todo for peer with no relation",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().HasAnyRelation(ctx, peerID, userID).Return(false, nil)

					todoRepo := mock.NewTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create todo for peer with relation error",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().HasAnyRelation(ctx, peerID, userID).Return(false, assert.AnError)

					todoRepo := mock.NewTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoCreate,
		},
		{
			name: "create todo for self",
			args: args{
				ctx:  context.Background(),
				todo: testModel.NewTodo(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, todo).Return(nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.todo),
			}
			err := s.Create(tt.args.ctx, tt.args.todo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTodoService_Get(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	todo := testModel.NewTodo(userID, userID)

	type args struct {
		ctx context.Context
		id  model.ID
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *model.Todo
		wantErr error
	}{
		{
			name: "get todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(todo, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			want: todo,
		},
		{
			name: "get todo with no permission",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get todo with permission error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindRead,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(nil, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    mock.NewMockLicenseService(ctrl),
					}
				},
			},
			wantErr: ErrTodoGet,
		},
		{
			name: "get todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			wantErr: ErrTodoGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			todo, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_GetAll(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type args struct {
		ctx           context.Context
		offset, limit int
		completed     *bool
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, completed *bool, todos []*model.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    []*model.Todo
		wantErr error
	}{
		{
			name: "get all todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				offset:    0,
				limit:     10,
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, completed *bool, todos []*model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/GetAll", gomock.Len(0)).Return(ctx, span)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().GetByOwner(ctx, userID, offset, limit, completed).Return(todos, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: []*model.Todo{
				testModel.NewTodo(userID, userID),
				testModel.NewTodo(userID, userID),
			},
		},
		{
			name: "get all completed todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				offset:    0,
				limit:     10,
				completed: convert.ToPointer(true),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, completed *bool, todos []*model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/GetAll", gomock.Len(0)).Return(ctx, span)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().GetByOwner(ctx, userID, offset, limit, completed).Return(todos, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: []*model.Todo{
				testModel.NewTodo(userID, userID),
				testModel.NewTodo(userID, userID),
			},
		},
		{
			name: "get all active todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				offset:    0,
				limit:     10,
				completed: convert.ToPointer(false),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, completed *bool, todos []*model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/GetAll", gomock.Len(0)).Return(ctx, span)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().GetByOwner(ctx, userID, offset, limit, completed).Return(todos, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: []*model.Todo{
				testModel.NewTodo(userID, userID),
				testModel.NewTodo(userID, userID),
			},
		},
		{
			name: "get todos with no context user id",
			args: args{
				ctx:       context.Background(),
				offset:    0,
				limit:     10,
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ *bool, _ []*model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "get todos with error",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				offset:    0,
				limit:     10,
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, completed *bool, _ []*model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/GetAll", gomock.Len(0)).Return(ctx, span)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().GetByOwner(ctx, userID, offset, limit, completed).Return(nil, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			wantErr: ErrTodoGetAll,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.args.completed, tt.want),
			}
			todo, err := s.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit, tt.args.completed)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	todo := testModel.NewTodo(userID, userID)

	type args struct {
		ctx   context.Context
		id    model.ID
		patch map[string]any
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *model.Todo
		wantErr error
	}{
		{
			name: "update todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, todo *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Update(ctx, id, patch).Return(todo, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			want: todo,
		},
		{
			name: "update todo with no permission",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update todo with permission error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindWrite,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Update(ctx, id, patch).Return(nil, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoUpdate,
		},
		{
			name: "update todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoUpdate,
		},
		{
			name: "update todo with expired license",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update todo with license error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: map[string]any{
					"title": "title",
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ map[string]any, _ *model.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			todo, err := s.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	todo := testModel.NewTodo(userID, userID)

	type args struct {
		ctx context.Context
		id  model.ID
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *model.Todo
		wantErr error
	}{
		{
			name: "delete todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Delete(ctx, id).Return(nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			want: todo,
		},
		{
			name: "delete todo with no permission",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete todo with permission error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(false)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					permSvc := mock.NewPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, []model.PermissionKind{
						model.PermissionKindDelete,
					}).Return(true)

					todoRepo := mock.NewTodoRepository(ctrl)
					todoRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoDelete,
		},
		{
			name: "delete todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrTodoDelete,
		},
		{
			name: "delete todo with expired license",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete todo with license error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          mock.NewTodoRepository(ctrl),
						permissionService: mock.NewPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
