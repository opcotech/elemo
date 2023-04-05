package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
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
				},
			},
			want: &userService{
				baseService: &baseService{
					logger:   new(mock.Logger),
					tracer:   new(mock.Tracer),
					userRepo: new(mock.UserRepository),
				},
			},
		},
		{
			name: "new user service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithUserRepository(new(mock.UserRepository)),
				},
			},
			wantErr: ErrNoLogger,
		},
		{
			name: "new user service with no user repository",
			args: args{
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoUserRepository,
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
					}
				},
			},
			args: args{
				ctx:  context.Background(),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				user: testModel.NewUser(),
			},
			wantErr: ErrUserCreate,
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
				id:  model.MustNewID(model.UserIDType),
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
				id:  model.MustNewID(model.UserIDType),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.UserIDType),
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			want: testModel.NewUser(),
		},
		{
			name: "update user with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID, patch map[string]any, user *model.User) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Update", []trace.SpanStartOption(nil)).Return(ctx, span)

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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.UserIDType),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.UserIDType),
				patch: map[string]any{
					"email": "test2@example.com",
				},
			},
			wantErr: ErrUserUpdate,
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.UserIDType),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.UserIDType),
				force: true,
			},
		},
		{
			name: "delete user with invalid id",
			fields: fields{
				baseService: func(ctx context.Context, id model.ID) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.userService/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: new(mock.UserRepository),
					}
				},
			},
			args: args{
				ctx:   context.Background(),
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

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.UserIDType),
				force: false,
			},
			wantErr: ErrUserDelete,
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
					userRepo.On("Delete", ctx, id).Return(errors.New("error"))

					return &baseService{
						logger:   new(mock.Logger),
						tracer:   tracer,
						userRepo: userRepo,
					}
				},
			},
			args: args{
				ctx:   context.Background(),
				id:    model.MustNewID(model.UserIDType),
				force: true,
			},
			wantErr: ErrUserDelete,
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