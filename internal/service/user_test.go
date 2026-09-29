package service_test

import (
	"context"
	"testing"

	"github.com/opcotech/elemo/internal/entitlement"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func createUserOptsFromRepo(o repository.CreateUserOpts) service.CreateUserOpts {
	return service.CreateUserOpts{
		Username:  o.Username,
		Email:     o.Email,
		Password:  o.Password,
		Status:    o.Status,
		FirstName: o.FirstName,
		LastName:  o.LastName,
		Picture:   o.Picture,
		Title:     o.Title,
		Bio:       o.Bio,
		Phone:     o.Phone,
		Address:   o.Address,
		Links:     o.Links,
		Languages: o.Languages,
	}
}

func repoUserToService(u *repository.User) *service.User {
	if u == nil {
		return nil
	}
	return service.UserFromRepository(u)
}

func repoUsersToService(users []*repository.User) []*service.User {
	out := make([]*service.User, len(users))
	for i, u := range users {
		out[i] = repoUserToService(u)
	}
	return out
}

func TestNewUserService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.UserService, error)
		wantErr error
	}{
		{
			name: "new user service",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), entitlement.Unrestricted(), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new user service with no user repository",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(nil, mockrepo.NewMockUserTokenRepository(nil), entitlement.Unrestricted(), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoUserRepository,
		},
		{
			name: "new user service with no user token repository",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), nil, entitlement.Unrestricted(), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoUserTokenRepository,
		},
		{
			name: "new user service with invalid options",
			build: func(_ *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), entitlement.Unrestricted(), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new user service with no seat policy",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), nil, service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: entitlement.ErrNoSeatPolicy,
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

func TestUserService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, opts service.CreateUserOpts) service.UserService
	}
	type args struct {
		ctx  context.Context
		opts service.CreateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&repository.User{}, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
		},
		{
			name: "create user with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: service.CreateUserOpts{},
			},
			wantErr: service.ErrUserCreate,
		},

		{
			name: "create user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: service.ErrUserCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts)
			_, err := s.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) service.UserService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(gomock.Any(), id, repository.UserDetailProjection()).Return(user, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: repoUserToService(testModel.NewUser()),
		},
		{
			name: "get user with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: service.ErrUserGet,
		},
		{
			name: "get user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(gomock.Any(), id, repository.UserDetailProjection()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: service.ErrUserGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := testModel.NewUser()
			if tt.want != nil {
				tt.want = repoUserToService(u)
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, u)
			got, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) service.UserService
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(gomock.Any(), email, repository.UserDetailProjection()).Return(user, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.Background(),
				email: "email@example.com",
			},
			want: repoUserToService(testModel.NewUser()),
		},
		{
			name: "get user with invalid user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ string, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.Background(),
				email: "",
			},
			wantErr: service.ErrUserGet,
		},
		{
			name: "get user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(gomock.Any(), email, repository.UserDetailProjection()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: service.ErrUserGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := testModel.NewUser()
			if tt.want != nil {
				tt.want = repoUserToService(u)
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.email, u)
			got, err := s.GetByEmail(tt.args.ctx, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_List(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, users []*repository.User) service.UserService
	}
	type args struct {
		ctx  context.Context
		page service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.User]
		wantErr error
	}{
		{
			name: "list users",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, users []*repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/List", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().List(gomock.Any(), page, repository.UserListProjection()).Return(repository.Page[*repository.User]{Items: users}, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				page: service.CursorPage{Size: 10},
			},
		},
		{
			name: "list users without authenticated user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CursorPage, _ []*repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.Background(),
				page: service.CursorPage{Size: 10},
			},
			wantErr: service.ErrUserList,
		},
		{
			name: "list users with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CursorPage, _ []*repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.Background(),
				page: service.CursorPage{Size: -1},
			},
			wantErr: service.ErrUserList,
		},
		{
			name: "list users with oversized page",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CursorPage, _ []*repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.Background(),
				page: service.CursorPage{Size: repository.MaxPageSize + 1},
			},
			wantErr: service.ErrUserList,
		},
		{
			name: "list users with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, page service.CursorPage, _ []*repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/List", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().List(gomock.Any(), page, repository.UserListProjection()).Return(repository.Page[*repository.User]{}, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, model.MustNewID(model.ResourceTypeUser)),
				page: service.CursorPage{Size: 10},
			},
			wantErr: service.ErrUserList,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			users := []*repository.User{testModel.NewUser(), testModel.NewUser()}
			if tt.wantErr == nil {
				tt.want = service.Page[*service.User]{Items: repoUsersToService(users)}
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.page, users)
			got, err := s.List(tt.args.ctx, tt.args.page)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	otherUserID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateUserOpts, user *repository.User) service.UserService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateUserOpts, user *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Activate(gomock.Any(), id, gomock.Any(), repository.UnrestrictedActivation()).Return(user, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			want: repoUserToService(testModel.NewUser()),
		},
		{
			name: "update another user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUserID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update user with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				opts: service.UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrUserUpdate,
		},
		{
			name: "update user with empty patch",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(nil, repository.ErrNotFound)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: service.ErrUserUpdate,
		},
		{
			name: "update user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx: context.Background(),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email: optional.Some("test@example.com"),
				},
			},
			wantErr: service.ErrNoUser,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := testModel.NewUser()
			if tt.want != nil {
				tt.want = repoUserToService(u)
			}
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, u)
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.UserService
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		force bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "soft delete user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(new(repository.User), nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				force: false,
			},
		},
		{
			name: "force delete user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(gomock.Any(), id).Return(nil).Times(1)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				force: true,
			},
		},
		{
			name: "soft delete another user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "force delete another user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete user with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.ID{},
				force: false,
			},
			wantErr: service.ErrUserDelete,
		},
		{
			name: "soft delete user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				force: false,
			},
			wantErr: service.ErrUserDelete,
		},
		{
			name: "force delete user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(gomock.Any(), id).Return(assert.AnError).Times(1)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				force: true,
			},
			wantErr: service.ErrUserDelete,
		},
		{
			name: "soft delete user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "force delete user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
							entitlement.Unrestricted(),
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
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: service.ErrNoUser,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
