package repository

import (
	"context"
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedPermissionRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, permission *model.Permission) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, permission *model.Permission) PermissionRepository
	}
	type args struct {
		ctx        context.Context
		permission *model.Permission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Permission) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, permission *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Create(ctx, permission).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				permission: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
		},
		{
			name: "add new permission with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Permission) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, permission *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Create(ctx, permission).Return(ErrPermissionCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				permission: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrPermissionCreate,
		},
		{
			name: "add new permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Permission) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ *model.Permission) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				permission: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ *model.Permission) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ *model.Permission) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				permission: &model.Permission{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.permission),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.permission),
			}
			err := r.Create(tt.args.ctx, tt.args.permission)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedPermissionRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permission *model.Permission) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permission *model.Permission) PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Permission
		wantErr error
	}{
		{
			name: "get permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permission *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(permission, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			want: &model.Permission{
				ID:      model.MustNewID(model.ResourceTypePermission),
				Kind:    model.PermissionKindRead,
				Subject: model.MustNewID(model.ResourceTypeUser),
				Target:  model.MustNewID(model.ResourceTypeProject),
			},
		},
		{
			name: "get permission with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetBySubject(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permission by subject",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetBySubject(ctx, id).Return(permissions, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
		},
		{
			name: "get permission by subject with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetBySubject(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetBySubject(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetByTarget(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permission by target",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, permissions []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetByTarget(ctx, id).Return(permissions, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
		},
		{
			name: "get permission by target with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetByTarget(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetByTarget(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetBySubjectAndTarget(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, permissions []*model.Permission) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, permissions []*model.Permission) PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Permission
		wantErr error
	}{
		{
			name: "get permission for target",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, permissions []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetBySubjectAndTarget(ctx, subject, target).Return(permissions, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Permission{
				{
					ID:      model.MustNewID(model.ResourceTypePermission),
					Kind:    model.PermissionKindRead,
					Subject: model.MustNewID(model.ResourceTypeUser),
					Target:  model.MustNewID(model.ResourceTypeProject),
				},
			},
		},
		{
			name: "get permission for target with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ []*model.Permission) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, _ []*model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().GetBySubjectAndTarget(ctx, subject, target).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := r.GetBySubjectAndTarget(tt.args.ctx, tt.args.subject, tt.args.target)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID, kind model.PermissionKind) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, kind model.PermissionKind, permission *model.Permission) PermissionRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		kind model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Permission
		wantErr error
	}{
		{
			name: "update permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.PermissionKind) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, kind model.PermissionKind, permission *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Update(ctx, id, kind).Return(permission, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			want: &model.Permission{
				ID:      model.MustNewID(model.ResourceTypePermission),
				Kind:    model.PermissionKindRead,
				Subject: model.MustNewID(model.ResourceTypeUser),
				Target:  model.MustNewID(model.ResourceTypeProject),
			},
		},
		{
			name: "update permission with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.PermissionKind) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, kind model.PermissionKind, _ *model.Permission) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Update(ctx, id, kind).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.PermissionKind) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ model.PermissionKind, _ *model.Permission) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.PermissionKind) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ model.PermissionKind, _ *model.Permission) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.kind),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.kind, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.kind)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) PermissionRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
		},
		{
			name: "delete permission with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "delete permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) *redisBaseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, rolesKey).Return(rolesKeyResult)
					dbClient.EXPECT().Keys(ctx, usersKey).Return(usersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, rolesKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, usersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) PermissionRepository {
					return mock.NewPermissionRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedPermissionRepository_HasPermission(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasPermission bool) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
		kinds   []model.PermissionKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "has permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasPermission(ctx, subject, target, kinds).Return(hasPermission, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				kinds: []model.PermissionKind{
					model.PermissionKindRead,
					model.PermissionKindWrite,
				},
			},
			want: true,
		},
		{
			name: "has no permission",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasPermission(ctx, subject, target, kinds).Return(hasPermission, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				kinds: []model.PermissionKind{
					model.PermissionKindRead,
					model.PermissionKindWrite,
				},
			},
			want: false,
		},
		{
			name: "has permission with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, _ bool, kinds []model.PermissionKind) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasPermission(ctx, subject, target, kinds).Return(false, ErrPermissionRead)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				kinds: []model.PermissionKind{
					model.PermissionKindRead,
					model.PermissionKindWrite,
				},
			},
			wantErr: ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want, tt.args.kinds),
			}
			got, err := r.HasPermission(tt.args.ctx, tt.args.subject, tt.args.target, tt.args.kinds...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_HasAnyRelation(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasAnyRelation bool) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasAnyRelation bool) PermissionRepository
	}
	type args struct {
		ctx     context.Context
		subject model.ID
		target  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "has system role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasAnyRelation bool) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasAnyRelation(ctx, subject, target).Return(hasAnyRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: true,
		},
		{
			name: "has no system role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, hasAnyRelation bool) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasAnyRelation(ctx, subject, target).Return(hasAnyRelation, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: false,
		},
		{
			name: "has system role with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, subject, target model.ID, _ bool) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasAnyRelation(ctx, subject, target).Return(false, ErrPermissionRead)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := r.HasAnyRelation(tt.args.ctx, tt.args.subject, tt.args.target)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_HasSystemRole(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctrl *gomock.Controller, ctx context.Context, source model.ID, hasSystemRole bool) *redisBaseRepository
		permissionRepo func(ctrl *gomock.Controller, ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) PermissionRepository
	}
	type args struct {
		ctx    context.Context
		source model.ID
		roles  []model.SystemRole
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "has system role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasSystemRole(ctx, source, roles).Return(hasSystemRole, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeUser),
				roles: []model.SystemRole{
					model.SystemRoleOwner,
					model.SystemRoleSupport,
				},
			},
			want: true,
		},
		{
			name: "has no system role",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasSystemRole(ctx, source, roles).Return(hasSystemRole, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeUser),
				roles: []model.SystemRole{
					model.SystemRoleOwner,
					model.SystemRoleSupport,
				},
			},
			want: false,
		},
		{
			name: "has system role with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ bool) *redisBaseRepository {
					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &redisBaseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				permissionRepo: func(ctrl *gomock.Controller, ctx context.Context, source model.ID, _ bool, roles []model.SystemRole) PermissionRepository {
					repo := mock.NewPermissionRepository(ctrl)
					repo.EXPECT().HasSystemRole(ctx, source, roles).Return(false, ErrPermissionRead)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeUser),
				roles: []model.SystemRole{
					model.SystemRoleOwner,
					model.SystemRoleSupport,
				},
			},
			wantErr: ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.source, tt.want),
				permissionRepo: tt.fields.permissionRepo(ctrl, tt.args.ctx, tt.args.source, tt.want, tt.args.roles),
			}
			got, err := r.HasSystemRole(tt.args.ctx, tt.args.source, tt.args.roles...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
