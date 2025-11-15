package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedNamespaceRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, creatorID, organization model.ID, namespace *model.Namespace) *baseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) repository.NamespaceRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, creatorID, organization model.ID, namespace *model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Create(ctx, creatorID, organization, namespace).Return(repository.ErrNamespaceCreate)
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
			wantErr: repository.ErrNamespaceCreate,
		},
		{
			name: "add new namespace with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ model.ID, _ *model.Namespace) repository.NamespaceRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new namespace with organization cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ *model.Namespace) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")
					organizationKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					organizationKeyResult := new(redis.StringSliceCmd)
					organizationKeyResult.SetVal([]string{organizationKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyResult)
					dbClient.EXPECT().Keys(ctx, organizationKey).Return(organizationKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ model.ID, _ *model.Namespace) repository.NamespaceRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedNamespaceRepository{
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
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) repository.NamespaceRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(**model.Namespace); ok {
							*ptr = namespace
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, _ *model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _ *model.Namespace) repository.NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, namespace *model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(namespace, nil)
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
			var want *model.Namespace
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedNamespaceRepository{
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
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *baseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) repository.NamespaceRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespaces,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*[]*model.Namespace); ok {
							*ptr = namespaces
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, _ []*model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, _ []*model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().GetAll(ctx, organization, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, _ []*model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID, _, _ int, _ []*model.Namespace) repository.NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, organization model.ID, offset, limit int, namespaces []*model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", organization.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespaces,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, organization model.ID, offset, limit int, namespaces []*model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().GetAll(ctx, organization, offset, limit).Return(namespaces, nil)
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
			r := &CachedNamespaceRepository{
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
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository
		namespaceRepo func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) repository.NamespaceRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Namespace) *baseRepository {
					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  mock.NewCacheBackend(ctrl),
						tracer: mock.NewMockTracer(ctrl),
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, _ *model.Namespace) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
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
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update namespace set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)
					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) repository.NamespaceRepository {
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update namespace delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, namespace *model.Namespace) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					cacheRepo := mock.NewCacheBackend(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: namespace,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID, patch map[string]any, namespace *model.Namespace) repository.NamespaceRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedNamespaceRepository{
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
		cacheRepo     func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) repository.NamespaceRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(ctx context.Context, ctrl *gomock.Controller, id model.ID) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					repo := mock.NewNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
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

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, organizationKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				namespaceRepo: func(_ context.Context, _ *gomock.Controller, _ model.ID) repository.NamespaceRepository {
					return mock.NewNamespaceRepository(nil)
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
			r := &CachedNamespaceRepository{
				cacheRepo:     tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				namespaceRepo: tt.fields.namespaceRepo(tt.args.ctx, ctrl, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
