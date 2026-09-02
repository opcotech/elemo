package service_test

import (
	"context"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

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
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateTodoOpts(owner, creator model.ID) service.CreateTodoOpts {
	return service.CreateTodoOpts{
		Title:       "test title",
		Description: "test description",
		Priority:    model.TodoPriorityNormal,
		Completed:   false,
		OwnedBy:     owner,
		CreatedBy:   creator,
		DueDate:     convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
	}
}

func newServiceTodo(owner, creator model.ID) *service.Todo {
	return service.TodoFromRepository(testModel.NewRepositoryTodo(owner, creator))
}

func TestNewTodoService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.TodoService, error)
		wantErr error
	}{
		{
			name: "new todo service",
			build: func(ctrl *gomock.Controller) (service.TodoService, error) {
				return service.NewTodoService(mockrepo.NewMockTodoRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new todo service with no todo repository",
			build: func(ctrl *gomock.Controller) (service.TodoService, error) {
				return service.NewTodoService(nil, mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoTodoRepository,
		},
		{
			name: "new todo service with no license service",
			build: func(ctrl *gomock.Controller) (service.TodoService, error) {
				return service.NewTodoService(mockrepo.NewMockTodoRepository(nil), nil, service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new todo service with invalid options",
			build: func(_ *gomock.Controller) (service.TodoService, error) {
				return service.NewTodoService(mockrepo.NewMockTodoRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func TestTodoService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeTodo)
	peerID := model.MustNewID(model.ResourceTypeTodo)

	type args struct {
		ctx  context.Context
		todo service.CreateTodoOpts
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, todo service.CreateTodoOpts) service.TodoService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryTodo(model.MustNewID(model.ResourceTypeUser), model.MustNewID(model.ResourceTypeUser)), nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create todo with invalid todo",
			args: args{
				ctx:  context.Background(),
				todo: service.CreateTodoOpts{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoCreate,
		},
		{
			name: "create todo with expired license",
			args: args{
				ctx:  context.Background(),
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoCreate,
		},
		{
			name: "create todo for peer with no relation",
			args: args{
				ctx:  context.Background(),
				todo: newCreateTodoOpts(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create todo for peer with relation error",
			args: args{
				ctx:  context.Background(),
				todo: newCreateTodoOpts(userID, peerID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoCreate,
		},
		{
			name: "create todo for self",
			args: args{
				ctx:  context.Background(),
				todo: newCreateTodoOpts(userID, userID),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateTodoOpts) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Create(ctx, gomock.Any()).Return(testModel.NewRepositoryTodo(model.MustNewID(model.ResourceTypeUser), model.MustNewID(model.ResourceTypeUser)), nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.todo)
			_, err := s.Create(tt.args.ctx, tt.args.todo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTodoService_Get(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	repoTodo := testModel.NewRepositoryTodo(userID, userID)
	todo := service.TodoFromRepository(repoTodo)

	type args struct {
		ctx context.Context
		id  model.ID
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) service.TodoService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *service.Todo
		wantErr error
	}{
		{
			name: "get todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, todo *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(todo, nil).Times(1)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "get todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(nil, assert.AnError).Times(1)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(ctrl),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoGet,
		},
		{
			name: "get todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			var repoT *repository.Todo
			if tt.want != nil {
				repoT = &repository.Todo{
					ID: tt.want.ID, Title: tt.want.Title, Description: tt.want.Description,
					Priority: tt.want.Priority, Completed: tt.want.Completed,
					OwnedBy: tt.want.OwnedBy, CreatedBy: tt.want.CreatedBy,
					DueDate: tt.want.DueDate, CreatedAt: tt.want.CreatedAt, UpdatedAt: tt.want.UpdatedAt,
				}
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, repoT)
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
		page      service.CursorPage
		completed *bool
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, completed *bool, todos []*repository.Todo) service.TodoService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    service.Page[*service.Todo]
		wantErr error
	}{
		{
			name: "get all todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      service.CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, completed *bool, todos []*repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			want: service.Page[*service.Todo]{Items: []*service.Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get all completed todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      service.CursorPage{Size: 10},
				completed: convert.ToPointer(true),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, completed *bool, todos []*repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			want: service.Page[*service.Todo]{Items: []*service.Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get all active todos",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      service.CursorPage{Size: 10},
				completed: convert.ToPointer(false),
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, completed *bool, todos []*repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{Items: todos}, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			want: service.Page[*service.Todo]{Items: []*service.Todo{
				newServiceTodo(userID, userID),
				newServiceTodo(userID, userID),
			}},
		},
		{
			name: "get todos with no context user id",
			args: args{
				ctx:       context.Background(),
				page:      service.CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CursorPage, _ *bool, _ []*repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "get todos with error",
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				page:      service.CursorPage{Size: 10},
				completed: nil,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, completed *bool, _ []*repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/List", gomock.Len(0)).Return(ctx, span)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().ListByOwner(ctx, userID, page, completed).Return(repository.Page[*repository.Todo]{}, assert.AnError)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							mocksvc.NewMockLicenseService(nil),
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoList,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repoTodos := make([]*repository.Todo, len(tt.want.Items))
			for i, w := range tt.want.Items {
				repoTodos[i] = &repository.Todo{
					ID: w.ID, Title: w.Title, Description: w.Description,
					Priority: w.Priority, Completed: w.Completed,
					OwnedBy: w.OwnedBy, CreatedBy: w.CreatedBy,
					DueDate: w.DueDate, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
				}
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.page, tt.args.completed, repoTodos)
			todo, err := s.List(tt.args.ctx, tt.args.page, tt.args.completed)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, todo)
		})
	}
}

func TestTodoService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	repoTodo := testModel.NewRepositoryTodo(userID, userID)
	todo := service.TodoFromRepository(repoTodo)

	type args struct {
		ctx   context.Context
		id    model.ID
		patch service.UpdateTodoOpts
	}
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch service.UpdateTodoOpts, todo *repository.Todo) service.TodoService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *service.Todo
		wantErr error
	}{
		{
			name: "update todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateTodoOpts, todo *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(todo, nil).Times(1)
					todoRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(todo, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			want: todo,
		},
		{
			name: "update todo owned by another user",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "update todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoUpdate,
		},
		{
			name: "update todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoUpdate,
		},
		{
			name: "update todo with expired license",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update todo with license error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
				patch: service.UpdateTodoOpts{
					Title: optional.Some("title"),
				},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateTodoOpts, _ *repository.Todo) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			var repoT *repository.Todo
			if tt.want != nil {
				repoT = &repository.Todo{
					ID: tt.want.ID, Title: tt.want.Title, Description: tt.want.Description,
					Priority: tt.want.Priority, Completed: tt.want.Completed,
					OwnedBy: tt.want.OwnedBy, CreatedBy: tt.want.CreatedBy,
					DueDate: tt.want.DueDate, CreatedAt: tt.want.CreatedAt, UpdatedAt: tt.want.UpdatedAt,
				}
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, repoT)
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
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.TodoService
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *service.Todo
		wantErr error
	}{
		{
			name: "delete todo",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Delete(ctx, id).Return(nil).Times(1)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					peerID := model.MustNewID(model.ResourceTypeUser)
					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(peerID, peerID), nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete todo with no user",
			args: args{
				ctx: context.Background(),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "delete todo with error",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					todoRepo := mockrepo.NewMockTodoRepository(ctrl)
					todoRepo.EXPECT().Get(ctx, id).Return(testModel.NewRepositoryTodo(userID, userID), nil)
					todoRepo.EXPECT().Delete(ctx, id).Return(assert.AnError).Times(1)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							todoRepo,
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoDelete,
		},
		{
			name: "delete todo with invalid id",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
				},
			},
			wantErr: service.ErrTodoDelete,
		},
		{
			name: "delete todo with expired license",
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  todo.ID,
			},
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.TodoService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.todoService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.TodoService {
						svc, err := service.NewTodoService(
							mockrepo.NewMockTodoRepository(ctrl),
							licenseSvc,
							service.WithLogger(mocklog.NewMockLogger(ctrl)),
							service.WithTracer(tracer),
						)
						if err != nil {
							panic(err)
						}
						return svc
					}()
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
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)
			err := s.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
