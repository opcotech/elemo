package repository_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

func mustNamespaceListForOrganizationKey(t *testing.T, orgID model.ID, page repository.CursorPage) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.NamespaceListQuery{
		OrgID:      orgID,
		ActorID:    model.MustNewNilID(model.ResourceTypeUser),
		Page:       page,
		Order:      repository.SortDirectionDesc,
		Projection: repository.NamespaceListProjection(),
	}, model.ResourceTypeNamespace.String(), "ListForOrganization", orgID.String())
}

func mustNamespaceListAccessibleKey(t *testing.T, actor model.ID, page repository.CursorPage) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.NamespaceListAccessibleQuery{
		ActorID:    actor,
		Page:       page,
		Order:      repository.SortDirectionDesc,
		Projection: repository.NamespaceListProjection(),
	}, model.ResourceTypeNamespace.String(), "ListAccessible")
}

func TestCachedNamespaceRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateNamespaceOpts) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNamespaceOpts) repository.NamespaceRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateNamespaceOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateNamespaceOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNamespaceOpts) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Namespace{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNamespaceOpts{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateNamespaceOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, opts repository.CreateNamespaceOpts) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrNamespaceCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: repository.ErrNamespaceCreate,
		},
		{
			name: "add new namespace with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateNamespaceOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ repository.CreateNamespaceOpts) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new namespace with organization cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateNamespaceOpts) []repository.RedisRepositoryOption {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					listAccessibleKeyResult := new(redis.StringSliceCmd)
					listAccessibleKeyResult.SetVal([]string{listAccessibleKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ repository.CreateNamespaceOpts) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateNamespaceOpts{
					Name:        "test namespace",
					Description: "test description",
					CreatorID:   model.MustNewID(model.ResourceTypeUser),
					OrgID:       model.MustNewID(model.ResourceTypeOrganization),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.opts),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedNamespaceRepository_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *repository.Namespace) repository.NamespaceRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Namespace
		wantErr error
	}{
		{
			name: "get uncached namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *repository.Namespace {
				return &repository.Namespace{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**repository.Namespace); ok {
							*ptr = namespace
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *repository.Namespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			want: func(id model.ID) *repository.Namespace {
				return &repository.Namespace{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, _ *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached namespace error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *repository.Namespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached namespace cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.NamespaceDetailProjection()).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *repository.Namespace
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.NamespaceDetailProjection())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedNamespaceRepository_List(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*repository.Namespace) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*repository.Namespace) repository.NamespaceRepository
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
		want    []*repository.Namespace
		wantErr error
	}{
		{
			name: "get uncached namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*repository.Namespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListForOrganizationKey(t, organization, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Namespace]{Items: namespaces},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListForOrganization(ctx, repository.NamespaceListQuery{OrgID: organization, ActorID: model.MustNewNilID(model.ResourceTypeUser), Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.Namespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*repository.Namespace{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*repository.Namespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListForOrganizationKey(t, organization, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*repository.Page[*repository.Namespace]); ok {
							*ptr = repository.Page[*repository.Namespace]{Items: namespaces}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*repository.Namespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*repository.Namespace{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, _ []*repository.Namespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListForOrganizationKey(t, organization, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, _ []*repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListForOrganization(ctx, repository.NamespaceListQuery{OrgID: organization, ActorID: model.MustNewNilID(model.ResourceTypeUser), Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.Namespace]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get namespaces cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, _ []*repository.Namespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListForOrganizationKey(t, organization, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*repository.Namespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached namespaces cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, _, limit int, namespaces []*repository.Namespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListForOrganizationKey(t, organization, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Namespace]{Items: namespaces},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, _, limit int, namespaces []*repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListForOrganization(ctx, repository.NamespaceListQuery{OrgID: organization, ActorID: model.MustNewNilID(model.ResourceTypeUser), Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.Namespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:          context.Background(),
				organization: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.organization, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.organization, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForOrganization(tt.args.ctx, repository.NamespaceListQuery{OrgID: tt.args.organization, ActorID: model.MustNewNilID(model.ResourceTypeUser), Page: repository.CursorPage{Size: testPageSize(tt.args.limit)}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedNamespaceRepository_ListAccessible(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) repository.NamespaceRepository
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
		want    []*repository.AccessibleNamespace
		wantErr error
	}{
		{
			name: "get uncached accessible namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListAccessibleKey(t, actor, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.AccessibleNamespace]{Items: namespaces},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, repository.NamespaceListAccessibleQuery{ActorID: actor, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.AccessibleNamespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.AccessibleNamespace{
				{
					Namespace: repository.Namespace{
						ID:            model.MustNewID(model.ResourceTypeNamespace),
						Name:          "test namespace",
						Description:   "test description",
						ProjectCount:  convert.ToPointer(int64(0)),
						DocumentCount: convert.ToPointer(int64(0)),
					},
					Organization: repository.PartialOrganization{
						ID:   model.MustNewID(model.ResourceTypeOrganization),
						Name: "test organization",
					},
				},
			},
		},
		{
			name: "get cached accessible namespaces",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListAccessibleKey(t, actor, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*repository.Page[*repository.AccessibleNamespace]); ok {
							*ptr = repository.Page[*repository.AccessibleNamespace]{Items: namespaces}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ int, _ []*repository.AccessibleNamespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			want: []*repository.AccessibleNamespace{
				{
					Namespace: repository.Namespace{
						ID:            model.MustNewID(model.ResourceTypeNamespace),
						Name:          "test namespace",
						Description:   "test description",
						ProjectCount:  convert.ToPointer(int64(0)),
						DocumentCount: convert.ToPointer(int64(0)),
					},
					Organization: repository.PartialOrganization{
						ID:   model.MustNewID(model.ResourceTypeOrganization),
						Name: "test organization",
					},
				},
			},
		},
		{
			name: "get uncached accessible namespaces error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, _ []*repository.AccessibleNamespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListAccessibleKey(t, actor, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, _ []*repository.AccessibleNamespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, repository.NamespaceListAccessibleQuery{ActorID: actor, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.AccessibleNamespace]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get accessible namespaces cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, _ []*repository.AccessibleNamespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListAccessibleKey(t, actor, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ int, _ []*repository.AccessibleNamespace) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached accessible namespaces cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) []repository.RedisRepositoryOption {
					key := mustNamespaceListAccessibleKey(t, actor, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.AccessibleNamespace]{Items: namespaces},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, actor model.ID, limit int, namespaces []*repository.AccessibleNamespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().ListAccessible(ctx, repository.NamespaceListAccessibleQuery{ActorID: actor, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()}).Return(repository.Page[*repository.AccessibleNamespace]{Items: namespaces}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				actor: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.actor, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.actor, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListAccessible(tt.args.ctx, repository.NamespaceListAccessibleQuery{ActorID: tt.args.actor, Page: repository.CursorPage{Size: testPageSize(tt.args.limit)}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedNamespaceRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch repository.UpdateNamespaceOpts, namespace *repository.Namespace) repository.NamespaceRepository
	}
	type args struct {
		ctx   context.Context
		id    model.ID
		patch repository.UpdateNamespaceOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Namespace
		wantErr error
	}{
		{
			name: "update namespace",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
					namespaceGenKey := issueListNamespaceGenKey(id)
					projectionEpochKey := issueListProjectionEpochKey()

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(7)

					tracer := mocktrace.NewMockTracer(ctrl)
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

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch repository.UpdateNamespaceOpts, namespace *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: repository.UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			want: &repository.Namespace{
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Namespace) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch repository.UpdateNamespaceOpts, _ *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: repository.UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update namespace set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch repository.UpdateNamespaceOpts, namespace *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: repository.UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update namespace delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *repository.Namespace) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(repository.NamespaceDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch repository.UpdateNamespaceOpts, namespace *repository.Namespace) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(namespace, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
				patch: repository.UpdateNamespaceOpts{
					Name:        optional.Some("updated namespace"),
					Description: optional.Some("updated description"),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id, tt.args.patch, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCachedNamespaceRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID) repository.NamespaceRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
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

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, namespaceGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: namespaceGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectionEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectionEpochKey, Value: int64(1)}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
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

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(8)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)
					cacheRepo.EXPECT().Get(ctx, namespaceGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: namespaceGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectionEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectionEpochKey, Value: int64(1)}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrNamespaceDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrNamespaceDelete,
		},
		{
			name: "delete namespace with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					repo := mockrepo.NewMockNamespaceRepository(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete namespace with get all cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete namespace with organization cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListForOrganization", "*")
					listAccessibleKey := composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})
					listAccessibleKeyCmd := new(redis.StringSliceCmd)
					listAccessibleKeyCmd.SetVal([]string{listAccessibleKey})

					organizationKeyCmd := new(redis.StringSliceCmd)
					organizationKeyCmd.SetVal([]string{organizationKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, listAccessibleKey).Return(listAccessibleKeyCmd)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listAccessibleKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					return mockrepo.NewMockNamespaceRepository(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeNamespace),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedNamespaceRepository {
				r, err := repository.NewCachedNamespaceRepository(
					tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
