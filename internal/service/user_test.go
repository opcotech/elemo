package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNewUserService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    UserService
		wantErr error
	}{
		{
			name: "new user service",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionRepository(new(mock.PermissionRepository)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			want: &userService{
				baseService: &baseService{
					logger:         new(mock.Logger),
					tracer:         new(mock.Tracer),
					userRepo:       new(mock.UserRepository),
					permissionRepo: new(mock.PermissionRepository),
					licenseService: new(mock.LicenseService),
				},
			},
		},
		{
			name: "new user service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithUserRepository(new(mock.UserRepository)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new user service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoUserRepository,
		},
		{
			name: "new user service with no permission repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithLicenseService(new(mock.LicenseService)),
				},
			},
			wantErr: ErrNoPermissionRepository,
		},
		{
			name: "new user service with no license service",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
					WithUserRepository(new(mock.UserRepository)),
					WithPermissionRepository(new(mock.PermissionRepository)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUserService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Create(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctx context.Context, user *model.User) *baseService
	}
	type args struct {
		ctx  context.Context
		user *model.User
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
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Create", ctx, user).Return(nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, model.MustNewNilID(model.ResourceTypeUser), []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: testModel.NewUser(),
			},
		},
		{
			name: "create user with invalid user",
			fields: fields{
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: &model.User{},
			},
			wantErr: ErrUserCreate,
		},
		{
			name: "create user with error",
			fields: fields{
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Create", ctx, user).Return(errors.New("error"))

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, model.MustNewNilID(model.ResourceTypeUser), []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: testModel.NewUser(),
			},
			wantErr: ErrUserCreate,
		},
		{
			name: "create user out of quota",
			fields: fields{
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, model.MustNewNilID(model.ResourceTypeUser), []model.PermissionKind{
						model.PermissionKindCreate,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: testModel.NewUser(),
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "create user with expired license",
			fields: fields{
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: testModel.NewUser(),
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create user with license expired error",
			fields: fields{
				baseService: func(ctx context.Context, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Create", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, errors.New("error"))

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				user: testModel.NewUser(),
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.user),
			}
			err := s.Create(tt.args.ctx, tt.args.user)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserService_Get(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, id model.ID, user *model.User) *baseService
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Get", ctx, id).Return(user, nil)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: testModel.NewUser(),
		},
		{
			name: "get user with invalid user",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
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
				baseService: func(ctx context.Context, id model.ID, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Get", ctx, id).Return(nil, errors.New("error"))

					return &baseService{
						logger:   new(mock.Logger),
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
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, email string, user *model.User) *baseService
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.User
		wantErr error
	}{
		{
			name: "get user",
			fields: fields{
				baseService: func(ctx context.Context, email string, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetByEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("GetByEmail", ctx, email).Return(user, nil)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				email: "email@example.com",
			},
			want: testModel.NewUser(),
		},
		{
			name: "get user with invalid user",
			fields: fields{
				baseService: func(ctx context.Context, email string, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetByEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
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
				baseService: func(ctx context.Context, email string, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetByEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("error"))

					return &baseService{
						logger:   new(mock.Logger),
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
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.email, tt.want),
			}
			got, err := s.GetByEmail(tt.args.ctx, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_GetAll(t *testing.T) {
	type fields struct {
		baseService func(ctx context.Context, offset, limit int, users []*model.User) *baseService
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
		want    []*model.User
		wantErr error
	}{
		{
			name: "get all users user",
			fields: fields{
				baseService: func(ctx context.Context, offset, limit int, users []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("GetAll", ctx, offset, limit).Return(users, nil)

					return &baseService{
						logger:   new(mock.Logger),
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
			want: []*model.User{
				testModel.NewUser(),
				testModel.NewUser(),
			},
		},
		{
			name: "get all users with invalid offset",
			fields: fields{
				baseService: func(ctx context.Context, offset, limit int, users []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
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
				baseService: func(ctx context.Context, limit, offset int, users []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
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
				baseService: func(ctx context.Context, offset, limit int, user []*model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/GetAll", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("GetAll", ctx, offset, limit).Return(nil, errors.New("error"))

					return &baseService{
						logger:   new(mock.Logger),
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
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
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
		baseService func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.User
		wantErr error
	}{
		{
			name: "update user",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(user, nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, id, id, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.UserStatusActive.String(),
				},
			},
			want: testModel.NewUser(),
		},
		{
			name: "update user with no permission",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(user, nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, otherUserID, id, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUserID),
				id:  userID,
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrNoPermission,
		},
		{
			name: "update user with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  model.ID{},
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrUserUpdate,
		},
		{
			name: "update user with empty patch",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:   context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:    userID,
				patch: map[string]any{},
			},
			wantErr: ErrUserUpdate,
		},
		{
			name: "update user with error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(nil, errors.New("error"))

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, id, id, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrUserUpdate,
		},
		{
			name: "update user out of quota",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, id, id, []model.PermissionKind{
						model.PermissionKindWrite,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)
					licenseSvc.On("WithinThreshold", ctx, license.QuotaUsers).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: permRepo,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.UserStatusActive.String(),
				},
			},
			wantErr: ErrQuotaExceeded,
		},
		{
			name: "update user with no context user id",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  userID,
				patch: map[string]any{
					"email": "test@example.com",
				},
			},
			wantErr: ErrNoUser,
		},
		{
			name: "update user with expired license",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(true, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.UserStatusActive.String(),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update user with expired license error",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, errors.New("test error"))

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				id:  userID,
				patch: map[string]any{
					"email":  "test2@example.com",
					"status": model.UserStatusActive.String(),
				},
			},
			wantErr: license.ErrLicenseExpired,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)

	type fields struct {
		baseService func(ctx context.Context, id model.ID) *baseService
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status":   model.UserStatusDeleted.String(),
						"password": password.UnusablePassword,
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(new(model.User), nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Delete", ctx, id).Return(nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
			name: "soft delete user with no permission",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status":   model.UserStatusDeleted.String(),
						"password": password.UnusablePassword,
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(new(model.User), nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Delete", ctx, id).Return(nil)

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(false, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					patch := map[string]any{
						"status":   model.UserStatusDeleted.String(),
						"password": password.UnusablePassword,
					}

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Update", ctx, id, patch).Return(nil, errors.New("error"))

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					userRepo := new(mock.UserRepository)
					userRepo.On("Delete", ctx, id).Return(errors.New("error"))

					permRepo := new(mock.PermissionRepository)
					permRepo.On("HasPermission", ctx, userID, id, []model.PermissionKind{
						model.PermissionKindDelete,
						model.PermissionKindAll,
					}).Return(true, nil)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       userRepo,
						permissionRepo: permRepo,
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return().Twice()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
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
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					licenseSvc := new(mock.LicenseService)
					licenseSvc.On("Expired", ctx).Return(false, nil)

					return &baseService{
						logger:         new(mock.Logger),
						tracer:         tracer,
						userRepo:       new(mock.UserRepository),
						permissionRepo: new(mock.PermissionRepository),
						licenseService: licenseSvc,
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
			s := &userService{
				baseService: tt.fields.baseService(tt.args.ctx, tt.args.id),
			}
			err := s.Delete(tt.args.ctx, tt.args.id, tt.args.force)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
