package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedPermissionRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, permission *model.Permission) *baseRepository
		permissionRepo func(ctx context.Context, permission *model.Permission) repository.PermissionRepository
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
				cacheRepo: func(ctx context.Context, _ *model.Permission) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, permission *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Create", ctx, permission).Return(nil)
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
				cacheRepo: func(ctx context.Context, _ *model.Permission) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, permission *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Create", ctx, permission).Return(repository.ErrPermissionCreate)
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
			wantErr: repository.ErrPermissionCreate,
		},
		{
			name: "add new permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ *model.Permission) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ *model.Permission) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)
					cacheRepo.On("Delete", ctx, usersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.permission),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.permission),
			}
			err := r.Create(tt.args.ctx, tt.args.permission)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedPermissionRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, permission *model.Permission) *baseRepository
		permissionRepo func(ctx context.Context, id model.ID, permission *model.Permission) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _ model.ID, _ *model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, permission *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, id).Return(permission, nil)
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
				cacheRepo: func(_ context.Context, _ model.ID, _ *model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, _ *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetBySubject(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, permissions []*model.Permission) *baseRepository
		permissionRepo func(ctx context.Context, id model.ID, permissions []*model.Permission) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, permissions []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubject", ctx, id).Return(permissions, nil)
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
				cacheRepo: func(_ context.Context, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, _ []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubject", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetBySubject(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetByTarget(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, permissions []*model.Permission) *baseRepository
		permissionRepo func(ctx context.Context, id model.ID, permissions []*model.Permission) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, permissions []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetByTarget", ctx, id).Return(permissions, nil)
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
				cacheRepo: func(_ context.Context, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, _ []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetByTarget", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetByTarget(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_GetBySubjectAndTarget(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, subject, target model.ID, permissions []*model.Permission) *baseRepository
		permissionRepo func(ctx context.Context, subject, target model.ID, permissions []*model.Permission) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, permissions []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubjectAndTarget", ctx, subject, target).Return(permissions, nil)
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ []*model.Permission) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, _ []*model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("GetBySubjectAndTarget", ctx, subject, target).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := r.GetBySubjectAndTarget(tt.args.ctx, tt.args.subject, tt.args.target)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, kind model.PermissionKind) *baseRepository
		permissionRepo func(ctx context.Context, id model.ID, kind model.PermissionKind, permission *model.Permission) repository.PermissionRepository
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
				cacheRepo: func(ctx context.Context, _ model.ID, _ model.PermissionKind) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, kind model.PermissionKind, permission *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Update", ctx, id, kind).Return(permission, nil)
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
				cacheRepo: func(ctx context.Context, _ model.ID, _ model.PermissionKind) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID, kind model.PermissionKind, _ *model.Permission) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Update", ctx, id, kind).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ model.ID, _ model.PermissionKind) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ model.ID, _ model.PermissionKind, _ *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ model.ID, _ model.PermissionKind) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ model.ID, _ model.PermissionKind, _ *model.Permission) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypePermission),
				kind: model.PermissionKindWrite,
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.kind),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id, tt.args.kind, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.kind)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID) *baseRepository
		permissionRepo func(ctx context.Context, id model.ID) repository.PermissionRepository
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
				cacheRepo: func(ctx context.Context, _ model.ID) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, _ model.ID) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, id model.ID) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "delete permission with roles cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ model.ID) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ model.ID) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete permission with users cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, _ model.ID) *baseRepository {
					rolesKey := composeCacheKey(model.ResourceTypeRole.String(), "*")
					usersKey := composeCacheKey(model.ResourceTypeUser.String(), "*")

					rolesKeyResult := new(redis.StringSliceCmd)
					rolesKeyResult.SetVal([]string{rolesKey})

					usersKeyResult := new(redis.StringSliceCmd)
					usersKeyResult.SetVal([]string{usersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, rolesKey).Return(rolesKeyResult)
					dbClient.On("Keys", ctx, usersKey).Return(usersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, usersKey).Return(nil)
					cacheRepo.On("Delete", ctx, rolesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(_ context.Context, _ model.ID) repository.PermissionRepository {
					return new(mock.PermissionRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypePermission),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedPermissionRepository_HasPermission(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, subject, target model.ID, hasPermission bool) *baseRepository
		permissionRepo func(ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, kinds).Return(hasPermission, nil)
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasPermission bool, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, kinds).Return(hasPermission, nil)
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, _ bool, kinds []model.PermissionKind) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasPermission", ctx, subject, target, kinds).Return(false, repository.ErrPermissionRead)
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
			wantErr: repository.ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want, tt.args.kinds),
			}
			got, err := r.HasPermission(tt.args.ctx, tt.args.subject, tt.args.target, tt.args.kinds...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_HasAnyRelation(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, subject, target model.ID, hasAnyRelation bool) *baseRepository
		permissionRepo func(ctx context.Context, subject, target model.ID, hasAnyRelation bool) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasAnyRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(hasAnyRelation, nil)
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, hasAnyRelation bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(hasAnyRelation, nil)
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
				cacheRepo: func(_ context.Context, _, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, subject, target model.ID, _ bool) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasAnyRelation", ctx, subject, target).Return(false, repository.ErrPermissionRead)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: model.MustNewID(model.ResourceTypeUser),
				target:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.subject, tt.args.target, tt.want),
			}
			got, err := r.HasAnyRelation(tt.args.ctx, tt.args.subject, tt.args.target)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedPermissionRepository_HasSystemRole(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, source model.ID, hasSystemRole bool) *baseRepository
		permissionRepo func(ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) repository.PermissionRepository
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
				cacheRepo: func(_ context.Context, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, source, roles).Return(hasSystemRole, nil)
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
				cacheRepo: func(_ context.Context, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, source model.ID, hasSystemRole bool, roles []model.SystemRole) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, source, roles).Return(hasSystemRole, nil)
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
				cacheRepo: func(_ context.Context, _ model.ID, _ bool) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				permissionRepo: func(ctx context.Context, source model.ID, _ bool, roles []model.SystemRole) repository.PermissionRepository {
					repo := new(mock.PermissionRepository)
					repo.On("HasSystemRole", ctx, source, roles).Return(false, repository.ErrPermissionRead)
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
			wantErr: repository.ErrPermissionRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedPermissionRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.source, tt.want),
				permissionRepo: tt.fields.permissionRepo(tt.args.ctx, tt.args.source, tt.want, tt.args.roles),
			}
			got, err := r.HasSystemRole(tt.args.ctx, tt.args.source, tt.args.roles...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
