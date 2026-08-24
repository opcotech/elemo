package service_test

import (
	"context"
	"testing"

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
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new user service with no user repository",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(nil, mockrepo.NewMockUserTokenRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoUserRepository,
		},
		{
			name: "new user service with no user token repository",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), nil, mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoUserTokenRepository,
		},
		{
			name: "new user service with no license service",
			build: func(ctrl *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), nil, service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new user service with invalid options",
			build: func(_ *gomock.Controller) (service.UserService, error) {
				return service.NewUserService(mockrepo.NewMockUserRepository(nil), mockrepo.NewMockUserTokenRepository(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
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
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.User{}, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: service.ErrUserCreate,
		},
		{
			name: "create user out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "create user with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create user with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ service.CreateUserOpts) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
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
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, id, repository.UserDetailProjection()).Return(user, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, id, repository.UserDetailProjection()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			defer ctrl.Finish()
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
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(user, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email, repository.UserDetailProjection()).Return(nil, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			defer ctrl.Finish()
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
					tracer.EXPECT().Start(ctx, "service.userService/List", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().List(ctx, page, repository.UserListProjection()).Return(repository.Page[*repository.User]{Items: users}, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/List", gomock.Len(0)).Return(ctx, span)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/List", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().List(ctx, page, repository.UserListProjection()).Return(repository.Page[*repository.User]{}, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			defer ctrl.Finish()
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
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(user, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, repository.ErrNotFound)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			name: "update user out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "update user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx: context.Background(),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email: optional.Some("test@example.com"),
				},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "update user with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update user with expired license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateUserOpts, _ *repository.User) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: service.UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(new(repository.User), nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(ctx, id).Return(nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				force: true,
			},
		},
		{
			name: "delete user with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete user with license expired error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "soft delete another user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) service.UserService {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := mockrepo.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							userRepo,
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return func() service.UserService {
						svc, err := service.NewUserService(
							mockrepo.NewMockUserRepository(ctrl),
							mockrepo.NewMockUserTokenRepository(ctrl),
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
			defer ctrl.Finish()
			s := tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id)
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
