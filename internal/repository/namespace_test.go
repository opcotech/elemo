package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedNamespaceRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, creatorID, organization model.ID, namespace *model.Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) NamespaceRepository
	}
	type args struct {
		ctx          context.Context
		creatorID    model.ID
		organization model.ID
		namespace    *model.Namespace
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, creatorID, organization, namespace).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				creatorID:    model.MustNewID(model.ResourceTypeUser),
				organization: model.MustNewID(model.ResourceTypeOrganization),
				namespace: &model.Namespace{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
			},
		},
		{
			name: "add new namespace with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, creatorID, organization, namespace).Return(ErrNamespaceCreate)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				creatorID:    model.MustNewID(model.ResourceTypeUser),
				organization: model.MustNewID(model.ResourceTypeOrganization),
				namespace: &model.Namespace{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
			},
			wantErr: ErrNamespaceCreate,
		},
		{
			name: "add new namespace with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ model.ID, _ *model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				creatorID:    model.MustNewID(model.ResourceTypeUser),
				organization: model.MustNewID(model.ResourceTypeOrganization),
				namespace: &model.Namespace{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new namespace with organization cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ model.ID, _ *model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				creatorID:    model.MustNewID(model.ResourceTypeUser),
				organization: model.MustNewID(model.ResourceTypeOrganization),
				namespace: &model.Namespace{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
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
			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.creatorID, tt.args.organization, tt.args.namespace),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.creatorID, tt.args.organization, tt.args.namespace),
			}
			err := r.Create(tt.args.ctx, tt.args.creatorID, tt.args.organization, tt.args.namespace)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedNamespaceRepository_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) NamespaceRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Namespace
		wantErr error
	}{
		{
			name: "get uncached namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *model.Namespace {
				return &model.Namespace{
					ID:          id,
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				}
			},
		},
		{
			name: "get cached namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Namespace); ok {
							*ptr = namespace
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *model.Namespace {
				return &model.Namespace{
					ID:          id,
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				}
			},
		},
		{
			name: "get uncached namespace error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, _ *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached namespace error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached namespace cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *model.Namespace
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedNamespaceRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) NamespaceRepository
	}
	type args struct {
		ctx          context.Context
		organization model.ID
		offset       int
		limit        int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Namespace
		wantErr error
	}{
		{
			name: "get uncached namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespaces,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().GetAll(ctx, organization, offset, limit).Return(namespaces, nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Namespace{
				{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
			},
		},
		{
			name: "get cached namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*[]*model.Namespace); ok {
							*ptr = namespaces
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.Namespace{
				{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeNamespace),
					Name:        "test namespace",
					Description: "test description",
					Projects:    make([]*model.NamespaceProject, 0),
					Documents:   make([]*model.NamespaceDocument, 0),
				},
			},
		},
		{
			name: "get uncached namespaces error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, _ []*model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, _ []*model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().GetAll(ctx, organization, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get namespaces cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, _ []*model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*model.Namespace) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached namespaces cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespaces,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().GetAll(ctx, organization, offset, limit).Return(namespaces, nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.organization, tt.args.offset, tt.args.limit, tt.want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.organization, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.organization, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedNamespaceRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) NamespaceRepository
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
		want    *model.Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: map[string]any{
					"name":        "updated namespace",
					"description": "updated description",
				},
			},
			want: &model.Namespace{
				ID:          model.MustNewID(model.ResourceTypeNamespace),
				Name:        "test namespace",
				Description: "test description",
				Projects:    make([]*model.NamespaceProject, 0),
				Documents:   make([]*model.NamespaceDocument, 0),
			},
		},
		{
			name: "update namespace with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Namespace) *redisBaseRepository {
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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, _ *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: map[string]any{
					"name":        "updated namespace",
					"description": "updated description",
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update namespace set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)
					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: map[string]any{
					"name":        "updated namespace",
					"description": "updated description",
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update namespace delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: map[string]any{
					"name":        "updated namespace",
					"description": "updated description",
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

			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedNamespaceRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID) NamespaceRepository
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
			name: "delete namespace success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
		},
		{
			name: "delete namespace with namespace deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrNamespaceDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrNamespaceDelete,
		},
		{
			name: "delete namespace with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) NamespaceRepository {
					repo := mock.NewNamespaceRepository(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete namespace with get all cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete namespace with organization cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
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
			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
