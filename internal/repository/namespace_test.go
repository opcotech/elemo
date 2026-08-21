package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedNamespaceRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, opts CreateNamespaceOpts) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, opts CreateNamespaceOpts) NamespaceRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateNamespaceOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNamespaceOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, opts CreateNamespaceOpts) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&Namespace{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
		},
		{
			name: "add new namespace with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNamespaceOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, opts CreateNamespaceOpts) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, ErrNamespaceCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrNamespaceCreate,
		},
		{
			name: "add new namespace with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNamespaceOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")

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
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ CreateNamespaceOpts) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new namespace with organization cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ CreateNamespaceOpts) *redisBaseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ CreateNamespaceOpts) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
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
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.opts),
			}
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedNamespaceRepository_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *Namespace) NamespaceRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *Namespace
		wantErr error
	}{
		{
			name: "get uncached namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, NamespaceDetailProjection()).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *Namespace {
				return &Namespace{
					ID:            id,
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get cached namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
						if ptr, ok := dst.(**Namespace); ok {
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
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *Namespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *Namespace {
				return &Namespace{
					ID:            id,
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get uncached namespace error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, _ *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, NamespaceDetailProjection()).Return(nil, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *Namespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, NamespaceDetailProjection()).Return(namespace, nil)
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
			var want *Namespace
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id, NamespaceDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedNamespaceRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*Namespace) NamespaceRepository
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
		want    []*Namespace
		wantErr error
	}{
		{
			name: "get uncached namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", organization.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						Value: Page[*Namespace]{Items: namespaces},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().List(ctx, organization, model.MustNewNilID(model.ResourceTypeUser), CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*Namespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*Namespace{
				{
					ID:            model.MustNewID(model.ResourceTypeNamespace),
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:            model.MustNewID(model.ResourceTypeNamespace),
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", organization.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						if ptr, ok := dst.(*Page[*Namespace]); ok {
							*ptr = Page[*Namespace]{Items: namespaces}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*Namespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*Namespace{
				{
					ID:            model.MustNewID(model.ResourceTypeNamespace),
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				},
				{
					ID:            model.MustNewID(model.ResourceTypeNamespace),
					Name:          "test namespace",
					Description:   "test description",
					ProjectCount:  convert.ToPointer(int64(0)),
					DocumentCount: convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached namespaces error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, _ []*Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", organization.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, _ []*Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().List(ctx, organization, model.MustNewNilID(model.ResourceTypeUser), CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*Namespace]{}, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, _ []*Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", organization.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*Namespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", organization.String(), model.MustNewNilID(model.ResourceTypeUser).String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						Value: Page[*Namespace]{Items: namespaces},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().List(ctx, organization, model.MustNewNilID(model.ResourceTypeUser), CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*Namespace]{Items: namespaces}, nil)
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
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.organization, tt.args.offset, testPageSize(tt.args.limit), tt.want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.organization, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.List(tt.args.ctx, tt.args.organization, model.MustNewNilID(model.ResourceTypeUser), CursorPage{Size: testPageSize(tt.args.limit)}, NamespaceListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedNamespaceRepository_ListAccessible(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*AccessibleNamespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*AccessibleNamespace) NamespaceRepository
	}
	type args struct {
		ctx   context.Context
		actor model.ID
		limit int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*AccessibleNamespace
		wantErr error
	}{
		{
			name: "get uncached accessible namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*AccessibleNamespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", actor.String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						Value: Page[*AccessibleNamespace]{Items: namespaces},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*AccessibleNamespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, actor, CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*AccessibleNamespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*AccessibleNamespace{
				{
					Namespace: Namespace{
						ID:            model.MustNewID(model.ResourceTypeNamespace),
						Name:          "test namespace",
						Description:   "test description",
						ProjectCount:  convert.ToPointer(int64(0)),
						DocumentCount: convert.ToPointer(int64(0)),
					},
					Organization: PartialOrganization{
						ID:   model.MustNewID(model.ResourceTypeOrganization),
						Name: "test organization",
					},
				},
			},
		},
		{
			name: "get cached accessible namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*AccessibleNamespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", actor.String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						if ptr, ok := dst.(*Page[*AccessibleNamespace]); ok {
							*ptr = Page[*AccessibleNamespace]{Items: namespaces}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ int, _ []*AccessibleNamespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*AccessibleNamespace{
				{
					Namespace: Namespace{
						ID:            model.MustNewID(model.ResourceTypeNamespace),
						Name:          "test namespace",
						Description:   "test description",
						ProjectCount:  convert.ToPointer(int64(0)),
						DocumentCount: convert.ToPointer(int64(0)),
					},
					Organization: PartialOrganization{
						ID:   model.MustNewID(model.ResourceTypeOrganization),
						Name: "test organization",
					},
				},
			},
		},
		{
			name: "get uncached accessible namespaces error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, _ []*AccessibleNamespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", actor.String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, _ []*AccessibleNamespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, actor, CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*AccessibleNamespace]{}, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get accessible namespaces cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, _ []*AccessibleNamespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", actor.String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ int, _ []*AccessibleNamespace) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached accessible namespaces cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*AccessibleNamespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", actor.String(), projectionCacheValue(NamespaceListProjection()), "", limit)

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
						Value: Page[*AccessibleNamespace]{Items: namespaces},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*AccessibleNamespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, actor, CursorPage{Size: limit}, NamespaceListProjection()).Return(Page[*AccessibleNamespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
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
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.actor, testPageSize(tt.args.limit), tt.want),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.actor, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.ListAccessible(tt.args.ctx, tt.args.actor, CursorPage{Size: testPageSize(tt.args.limit)}, NamespaceListProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedNamespaceRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch UpdateNamespaceOpts, namespace *Namespace) NamespaceRepository
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch UpdateNamespaceOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					namespaceGenKey := issueListNamespaceGenKey(id)
					projectionEpochKey := issueListProjectionEpochKey()

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)

					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(7)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, namespaceGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: namespaceGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectionEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectionEpochKey, Value: int64(1)}).Return(nil)
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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch UpdateNamespaceOpts, namespace *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			want: &Namespace{
				ID:            model.MustNewID(model.ResourceTypeNamespace),
				Name:          "test namespace",
				Description:   "test description",
				ProjectCount:  convert.ToPointer(int64(0)),
				DocumentCount: convert.ToPointer(int64(0)),
			},
		},
		{
			name: "update namespace with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Namespace) *redisBaseRepository {
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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch UpdateNamespaceOpts, _ *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update namespace set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch UpdateNamespaceOpts, namespace *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update namespace delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *Namespace) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")

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
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch UpdateNamespaceOpts, namespace *Namespace) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
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
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					namespaceGenKey := issueListNamespaceGenKey(id)
					projectionEpochKey := issueListProjectionEpochKey()

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, namespaceGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: namespaceGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectionEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectionEpochKey, Value: int64(1)}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
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
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")
					namespaceGenKey := issueListNamespaceGenKey(id)
					projectionEpochKey := issueListProjectionEpochKey()

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, namespaceGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: namespaceGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectionEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectionEpochKey, Value: int64(1)}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) NamespaceRepository {
					repo := NewMockNamespaceRepository(ctrl)
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
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

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
					repo := NewMockNamespaceRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

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
					return NewMockNamespaceRepository(nil)
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
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) NamespaceRepository {
					return NewMockNamespaceRepository(nil)
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
