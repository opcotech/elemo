package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserOptsFromRepo(o repository.CreateUserOpts) CreateUserOpts {
	return CreateUserOpts{
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

func repoUserToService(u *repository.User) *User {
	if u == nil {
		return nil
	}
	return userFromRepository(u)
}

func repoUsersToService(users []*repository.User) []*User {
	out := make([]*User, len(users))
	for i, u := range users {
		out[i] = repoUserToService(u)
	}
	return out
}

func TestNewUserService(t *testing.T) {
	type args struct {
		opts func(ctrl *gomock.Controller) []Option
	}
	tests := []struct {
		name    string
		args    args
		want    func(ctrl *gomock.Controller) UserService
		wantErr error
	}{
		{
			name: "new user service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithUserRepository(repository.NewMockUserRepository(nil)),
						WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			want: func(ctrl *gomock.Controller) UserService {
				return &userService{
					baseService: &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            mock.NewMockTracer(ctrl),
						userRepo:          repository.NewMockUserRepository(nil),
						userTokenRepo:     repository.NewMockUserTokenRepository(nil),
						permissionService: NewMockPermissionService(nil),
						licenseService:    mock.NewMockLicenseService(nil),
					},
				}
			},
		},
		{
			name: "new user service with invalid options",
			args: args{
				opts: func(_ *gomock.Controller) []Option {
					return []Option{
						WithLogger(nil),
						WithUserRepository(repository.NewMockUserRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new user service with no user repository",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoUserRepository,
		},
		{
			name: "new user service with no permission repository",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithUserRepository(repository.NewMockUserRepository(nil)),
						WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
						WithLicenseService(mock.NewMockLicenseService(nil)),
					}
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new user service with no license service",
			args: args{
				opts: func(ctrl *gomock.Controller) []Option {
					return []Option{
						WithLogger(mock.NewMockLogger(ctrl)),
						WithTracer(mock.NewMockTracer(ctrl)),
						WithUserRepository(repository.NewMockUserRepository(nil)),
						WithUserTokenRepository(repository.NewMockUserTokenRepository(nil)),
						WithPermissionService(NewMockPermissionService(nil)),
					}
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := NewUserService(tt.args.opts(ctrl)...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.want != nil {
				assert.Equal(t, tt.want(ctrl), got)
			}
		})
	}
}

func TestUserService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, opts CreateUserOpts) *baseService
	}
	type args struct {
		ctx  context.Context
		opts CreateUserOpts
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(ctx, gomock.Any()).Return(&repository.User{}, nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						userRepo:       repository.NewMockUserRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: CreateUserOpts{},
			},
			wantErr: ErrUserCreate,
		},
		{
			name: "create user with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "create user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: ErrUserCreate,
		},
		{
			name: "create user out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				opts: createUserOptsFromRepo(testModel.NewCreateUserOpts()),
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create user with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ CreateUserOpts) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
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
			s := &userService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := s.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) *baseService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, user *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, id).Return(user, nil)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: repository.NewMockUserRepository(ctrl),
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: ErrUserGet,
		},
		{
			name: "get user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Get", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Get(ctx, id).Return(nil, assert.AnError)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrUserGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &userService{
				baseService: func() *baseService {
					u := testModel.NewUser()
					if tt.want != nil {
						tt.want = repoUserToService(u)
					}
					return tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, u)
				}(),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) *baseService
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, email string, user *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ string, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: repository.NewMockUserRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "",
			},
			wantErr: ErrUserGet,
		},
		{
			name: "get user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, email string, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetByEmail", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetByEmail(ctx, email).Return(nil, assert.AnError)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			wantErr: ErrUserGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &userService{
				baseService: func() *baseService {
					u := testModel.NewUser()
					if tt.want != nil {
						tt.want = repoUserToService(u)
					}
					return tt.fields.baseService(ctrl, tt.args.ctx, tt.args.email, u)
				}(),
			}
			got, err := s.GetByEmail(tt.args.ctx, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_GetAll(t *testing.T) {
	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*repository.User) *baseService
	}
	type args struct {
		ctx    context.Context
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*User
		wantErr error
	}{
		{
			name: "get all users user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, users []*repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetAll", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetAll(ctx, offset, limit).Return(users, nil)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*User{ // converted
				repoUserToService(testModel.NewUser()),
				repoUserToService(testModel.NewUser()),
			},
		},
		{
			name: "get all users with invalid offset",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: repository.NewMockUserRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: -1,
				limit:  10,
			},
			wantErr: ErrUserGetAll,
		},
		{
			name: "get all users with invalid limit",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _, _ int, _ []*repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetAll", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: repository.NewMockUserRepository(ctrl),
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  -1,
			},
			wantErr: ErrUserGetAll,
		},
		{
			name: "get all users with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/GetAll", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().GetAll(ctx, offset, limit).Return(nil, assert.AnError)

					return &baseService{
						logger:   mock.NewMockLogger(ctrl),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrUserGetAll,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &userService{
				baseService: func() *baseService {
					users := []*repository.User{testModel.NewUser(), testModel.NewUser()}
					if tt.want != nil {
						tt.want = repoUsersToService(users)
					}
					return tt.fields.baseService(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, users)
				}(),
			}
			got, err := s.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	otherUserID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateUserOpts, user *repository.User) *baseService
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateUserOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateUserOpts, user *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(user, nil)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						userRepo:       userRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			want: repoUserToService(testModel.NewUser()),
		},
		{
			name: "update user with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindWrite).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUserID),
				id:  userID,
				opts: UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update user with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						userRepo:       repository.NewMockUserRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				opts: UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrUserUpdate,
		},
		{
			name: "update user with empty patch",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, repository.ErrNotFound)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: UpdateUserOpts{
					Email: optional.Some("test2@example.com"),
				},
			},
			wantErr: ErrUserUpdate,
		},
		{
			name: "update user out of quota",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaUsers).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						userRepo:       repository.NewMockUserRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "update user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:         mock.NewMockLogger(ctrl),
						tracer:         tracer,
						userRepo:       repository.NewMockUserRepository(ctrl),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  userID,
				opts: UpdateUserOpts{
					Email: optional.Some("test@example.com"),
				},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "update user with expired license",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: UpdateUserOpts{
					Email:  optional.Some("test2@example.com"),
					Status: optional.Some(model.UserStatusActive),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update user with expired license error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ UpdateUserOpts, _ *repository.User) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				opts: UpdateUserOpts{
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
			s := &userService{
				baseService: func() *baseService {
					u := testModel.NewUser()
					if tt.want != nil {
						tt.want = repoUserToService(u)
					}
					return tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, u)
				}(),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(new(repository.User), nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
		},
		{
			name: "force delete user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
		},
		{
			name: "delete user with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, assert.AnError)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
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
			name: "soft delete user with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "force delete user with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(false)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "delete user with invalid id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.ID{},
				force: false,
			},
			wantErr: ErrUserDelete,
		},
		{
			name: "soft delete user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Update(ctx, id, gomock.Any()).Return(nil, assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
			wantErr: ErrUserDelete,
		},
		{
			name: "force delete user with error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					userRepo := repository.NewMockUserRepository(ctrl)
					userRepo.EXPECT().Delete(ctx, id).Return(assert.AnError)

					permSvc := NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserHasPermission(ctx, id, model.PermissionKindDelete).Return(true)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          userRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: ErrUserDelete,
		},
		{
			name: "soft delete user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: false,
			},
			wantErr: ErrNoUser,
		},
		{
			name: "force delete user with no context user id",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.userService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mock.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return &baseService{
						logger:            mock.NewMockLogger(ctrl),
						tracer:            tracer,
						userRepo:          repository.NewMockUserRepository(ctrl),
						permissionService: NewMockPermissionService(ctrl),
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.ResourceTypeUser),
				force: true,
			},
			wantErr: ErrNoUser,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := &userService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
