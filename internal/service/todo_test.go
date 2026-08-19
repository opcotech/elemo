package service

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateTodoOpts(owner, creator model.ID) CreateTodoOpts {
	return CreateTodoOpts{
		Title:       "test title",
		Description: "test description",
		Priority:    model.TodoPriorityNormal,
		Completed:   false,
		OwnedBy:     owner,
		CreatedBy:   creator,
		DueDate:     convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
	}
}

func newServiceTodo(owner, creator model.ID) *Todo {
	return todoFromRepository(testModel.NewRepositoryTodo(owner, creator))
}

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
					WithTodoRepository(repository.NewMockTodoRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &todoService{
				baseService: &baseService{
					logger:         mock.NewMockLogger(nil),
					tracer:         mock.NewMockTracer(nil),
					todoRepo:       repository.NewMockTodoRepository(nil),
					licenseService: mock.NewMockLicenseService(nil),
				},
			},
		},
		{
			name: "new todo service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTodoRepository(repository.NewMockTodoRepository(nil)),
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
			name: "new todo service with no license service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTodoRepository(repository.NewMockTodoRepository(nil)),
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
		todo CreateTodoOpts
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, todo CreateTodoOpts) *baseService
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
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryTodo(model.MustNewID(model.ResourceTypeUser), model.MustNewID(model.ResourceTypeUser)), nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
					}
				},
			},
		},
		{
			name: "create todo for peer",
			args: args{
				ctx:  context.Background(),
				todo: newCreateTodoOpts(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create todo with invalid todo",
			args: args{
				ctx:  context.Background(),
				todo: CreateTodoOpts{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					todoRepo := repository.NewMockTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateTodoOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryTodo(model.MustNewID(model.ResourceTypeUser), model.MustNewID(model.ResourceTypeUser)), nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
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
			_, err := s.Create(tt.args.ctx, tt.args.todo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTodoService_Get(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	repoTodo := testModel.NewRepositoryTodo(userID, userID)
	todo := todoFromRepository(repoTodo)

	type args struct {
		ctx context.Context
		id  model.ID
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *Todo
		wantErr error
	}{
		{
			name: "get todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(todo, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
			},
			want: todo,
		},
		{
			name: "get todo owned by another user",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "get todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       repository.NewMockTodoRepository(ctrl),
						licenseService: mock.NewMockLicenseService(ctrl),
					}
				},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "get todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(nil, assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: mock.NewMockLicenseService(ctrl),
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
			var repoT *repository.Todo
			if tt.want != nil {
				repoT = &repository.Todo{
					ID: tt.want.ID, Title: tt.want.Title, Description: tt.want.Description,
					Priority: tt.want.Priority, Completed: tt.want.Completed,
					OwnedBy: tt.want.OwnedBy, CreatedBy: tt.want.CreatedBy,
					DueDate: tt.want.DueDate, CreatedAt: tt.want.CreatedAt, UpdatedAt: tt.want.UpdatedAt,
				}
			}
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, repoT),
			}
			todo, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_List(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type args struct {
		ctx       context.Context
		page      CursorPage
		completed *bool
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, page CursorPage, completed *bool, todos []*repository.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    Page[*Todo]
		wantErr error
	}{
		{
			name: "get all todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page CursorPage, completed *bool, todos []*repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: Page[*Todo]{Items: []*Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get all completed todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      CursorPage{Size: 10},
				completed: convert.ToPointer(true),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page CursorPage, completed *bool, todos []*repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: Page[*Todo]{Items: []*Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get all active todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      CursorPage{Size: 10},
				completed: convert.ToPointer(false),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page CursorPage, completed *bool, todos []*repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    mock.NewMockLicenseService(nil),
					}
				},
			},
			want: Page[*Todo]{Items: []*Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get todos with no context user id",
			args: args{
				ctx:       context.Background(),
				page:      CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CursorPage, _ *bool, _ []*repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
				page:      CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page CursorPage, completed *bool, _ []*repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{}, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          todoRepo,
						permissionService: NewMockPermissionService(ctrl),
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
			repoTodos := make([]*repository.Todo, len(tt.want.Items))
			for i, w := range tt.want.Items {
				repoTodos[i] = &repository.Todo{
					ID: w.ID, Title: w.Title, Description: w.Description,
					Priority: w.Priority, Completed: w.Completed,
					OwnedBy: w.OwnedBy, CreatedBy: w.CreatedBy,
					DueDate: w.DueDate, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
				}
			}
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.page, tt.args.completed, repoTodos),
			}
			todo, err := s.List(tt.args.ctx, tt.args.page, tt.args.completed)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	repoTodo := testModel.NewRepositoryTodo(userID, userID)
	todo := todoFromRepository(repoTodo)

	type args struct {
		ctx   context.Context
		id    model.ID
		patch UpdateTodoOpts
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch UpdateTodoOpts, todo *repository.Todo) *baseService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *Todo
		wantErr error
	}{
		{
			name: "update todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateTodoOpts, todo *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(todo, nil)
					todoRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(todo, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
					}
				},
			},
			want: todo,
		},
		{
			name: "update todo owned by another user",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       repository.NewMockTodoRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "update todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
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
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
				patch: UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateTodoOpts, _ *repository.Todo) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
			var repoT *repository.Todo
			if tt.want != nil {
				repoT = &repository.Todo{
					ID: tt.want.ID, Title: tt.want.Title, Description: tt.want.Description,
					Priority: tt.want.Priority, Completed: tt.want.Completed,
					OwnedBy: tt.want.OwnedBy, CreatedBy: tt.want.CreatedBy,
					DueDate: tt.want.DueDate, CreatedAt: tt.want.CreatedAt, UpdatedAt: tt.want.UpdatedAt,
				}
			}
			s := &todoService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, repoT),
			}
			todo, err := s.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	todo := newServiceTodo(userID, userID)

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
		want    *Todo
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

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Delete(ctx, id).Return(nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
					}
				},
			},
			want: todo,
		},
		{
			name: "delete todo owned by another user",
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

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
					}
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
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
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       repository.NewMockTodoRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			wantErr: ErrNoUser,
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

					todoRepo := repository.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						todoRepo:       todoRepo,
						licenseService: licenseSvc,
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
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
						todoRepo:          repository.NewMockTodoRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
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
