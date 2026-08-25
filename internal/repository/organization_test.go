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

func mustOrganizationListForUserKey(t *testing.T, userID model.ID, page repository.CursorPage) string {
	t.Helper()
	return mustPlanCacheKey(t, repository.OrganizationListQuery{
		UserID:     userID,
		Action:     model.ActionOrganizationRead,
		Page:       page,
		Order:      repository.SortDirectionDesc,
		Projection: repository.OrganizationListProjection(),
	}, model.ResourceTypeOrganization.String(), "ListForUser")
}

func TestCachedOrganizationRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateOrganizationOpts) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateOrganizationOpts) repository.OrganizationRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateOrganizationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add new organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateOrganizationOpts) []repository.RedisRepositoryOption {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")
					refKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})
					refKeyResult := new(redis.StringSliceCmd)
					refKeyResult.SetVal([]string{refKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)
					dbClient.EXPECT().Keys(ctx, refKey).Return(refKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, refKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateOrganizationOpts) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Organization{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateOrganizationOpts{
					Slug:    "acme-org",
					Owner:   model.MustNewID(model.ResourceTypeUser),
					Name:    "test organization",
					Email:   "info@example.com",
					Logo:    "https://example.com/logo.png",
					Website: "https://example.com",
					Status:  model.OrganizationStatusActive,
				},
			},
		},
		{
			name: "add new organization with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateOrganizationOpts) []repository.RedisRepositoryOption {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")
					refKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})
					refKeyResult := new(redis.StringSliceCmd)
					refKeyResult.SetVal([]string{refKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)
					dbClient.EXPECT().Keys(ctx, refKey).Return(refKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, refKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateOrganizationOpts) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrOrganizationCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateOrganizationOpts{
					Slug:    "acme-org",
					Owner:   model.MustNewID(model.ResourceTypeUser),
					Name:    "test organization",
					Email:   "info@example.com",
					Logo:    "https://example.com/logo.png",
					Website: "https://example.com",
					Status:  model.OrganizationStatusActive,
				},
			},
			wantErr: repository.ErrOrganizationCreate,
		},
		{
			name: "add new organization get all cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ repository.CreateOrganizationOpts) []repository.RedisRepositoryOption {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateOrganizationOpts) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateOrganizationOpts{
					Slug:    "acme-org",
					Owner:   model.MustNewID(model.ResourceTypeUser),
					Name:    "test organization",
					Email:   "info@example.com",
					Logo:    "https://example.com/logo.png",
					Website: "https://example.com",
					Status:  model.OrganizationStatusActive,
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedOrganizationRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) repository.OrganizationRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Organization
		wantErr error
	}{
		{
			name: "get uncached organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

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
						Value: organization,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *repository.Organization {
				return &repository.Organization{
					ID:             id,
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get cached organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

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
						if orgPtr, ok := dst.(**repository.Organization); ok {
							*orgPtr = organization
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Organization) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *repository.Organization {
				return &repository.Organization{
					ID:             id,
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get uncached organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Organization) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached organization cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

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
						Value: organization,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.OrganizationDetailProjection()).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
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
			var want *repository.Organization
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.OrganizationDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedOrganizationRepository_GetByRef(t *testing.T) {
	t.Parallel()

	slug := "acme-org"
	proj := repository.OrganizationDetailProjection()
	org := &repository.Organization{
		ID:   model.MustNewID(model.ResourceTypeOrganization),
		Slug: slug,
		Name: "ACME",
	}
	cacheKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", slug, projectionCacheValue(proj))

	t.Run("reads uncached organization by slug", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

		backend := mockrepo.NewMockCacheBackend(ctrl)
		backend.EXPECT().Get(ctx, cacheKey, gomock.Any()).Return(cache.ErrCacheMiss)
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: cacheKey, Value: org}).Return(nil)

		inner := mockrepo.NewMockOrganizationRepository(ctrl)
		inner.EXPECT().GetByRef(ctx, model.ID{}, slug, proj).Return(org, nil)

		repo, err := repository.NewCachedOrganizationRepository(inner,
			repository.WithRedisDatabase(db),
			repository.WithCacheBackend(backend),
			repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
			repository.WithRedisRepositoryTracer(tracer),
		)
		require.NoError(t, err)

		got, err := repo.GetByRef(ctx, model.ID{}, slug, proj)
		require.NoError(t, err)
		require.Equal(t, org, got)
	})

	t.Run("reads cached organization by slug", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

		backend := mockrepo.NewMockCacheBackend(ctrl)
		backend.EXPECT().Get(ctx, cacheKey, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
			if ptr, ok := dst.(**repository.Organization); ok {
				*ptr = org
			}
		}).Return(nil)

		repo, err := repository.NewCachedOrganizationRepository(mockrepo.NewMockOrganizationRepository(ctrl),
			repository.WithRedisDatabase(db),
			repository.WithCacheBackend(backend),
			repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
			repository.WithRedisRepositoryTracer(tracer),
		)
		require.NoError(t, err)

		got, err := repo.GetByRef(ctx, model.ID{}, slug, proj)
		require.NoError(t, err)
		require.Equal(t, org, got)
	})
}

func TestCachedOrganizationRepository_List(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) repository.OrganizationRepository
	}
	type args struct {
		ctx    context.Context
		userID model.ID
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.Organization
		wantErr error
	}{
		{
			name: "get uncached organizations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) []repository.RedisRepositoryOption {
					key := mustOrganizationListForUserKey(t, userID, repository.CursorPage{Size: limit})

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Organization]{Items: organizations},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{UserID: userID, Action: model.ActionOrganizationRead, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.OrganizationListProjection()}).Return(repository.Page[*repository.Organization]{Items: organizations}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*repository.Organization{
				{
					ID:             model.MustNewID(model.ResourceTypeOrganization),
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				},
				{
					ID:             model.MustNewID(model.ResourceTypeOrganization),
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get cached organizations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) []repository.RedisRepositoryOption {
					key := mustOrganizationListForUserKey(t, userID, repository.CursorPage{Size: limit})

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
						if ptr, ok := dst.(*repository.Page[*repository.Organization]); ok {
							*ptr = repository.Page[*repository.Organization]{Items: organizations}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Organization) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*repository.Organization{
				{
					ID:             model.MustNewID(model.ResourceTypeOrganization),
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				},
				{
					ID:             model.MustNewID(model.ResourceTypeOrganization),
					Name:           "test organization",
					Email:          "info@example.com",
					Logo:           "https://example.com/logo.png",
					Website:        "https://example.com",
					Status:         model.OrganizationStatusActive,
					NamespaceCount: convert.ToPointer(int64(0)),
					TeamCount:      convert.ToPointer(int64(0)),
					MemberCount:    convert.ToPointer(int64(0)),
				},
			},
		},
		{
			name: "get uncached organizations error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Organization) []repository.RedisRepositoryOption {
					key := mustOrganizationListForUserKey(t, userID, repository.CursorPage{Size: limit})

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{UserID: userID, Action: model.ActionOrganizationRead, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.OrganizationListProjection()}).Return(repository.Page[*repository.Organization]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get organizations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, _ []*repository.Organization) []repository.RedisRepositoryOption {
					key := mustOrganizationListForUserKey(t, userID, repository.CursorPage{Size: limit})

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Organization) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached organizations cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) []repository.RedisRepositoryOption {
					key := mustOrganizationListForUserKey(t, userID, repository.CursorPage{Size: limit})

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Organization]{Items: organizations},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, _, limit int, organizations []*repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().ListForUser(ctx, repository.OrganizationListQuery{UserID: userID, Action: model.ActionOrganizationRead, Page: repository.CursorPage{Size: limit}, Order: repository.SortDirectionDesc, Projection: repository.OrganizationListProjection()}).Return(repository.Page[*repository.Organization]{Items: organizations}, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForUser(tt.args.ctx, repository.OrganizationListQuery{UserID: tt.args.userID, Action: model.ActionOrganizationRead, Page: repository.CursorPage{Size: testPageSize(tt.args.limit)}, Order: repository.SortDirectionDesc, Projection: repository.OrganizationListProjection()})
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedOrganizationRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateOrganizationOpts, organization *repository.Organization) repository.OrganizationRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateOrganizationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Organization
		wantErr error
	}{
		{
			name: "update organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", organization.Slug, projectionCacheValue(repository.OrganizationDetailProjection())),
						Value: organization,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateOrganizationOpts, organization *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				opts: repository.UpdateOrganizationOpts{
					Name: optional.Some("updated organization"),
				},
			},
			want: &repository.Organization{
				ID:             model.MustNewID(model.ResourceTypeOrganization),
				Slug:           "acme-org",
				Name:           "test organization",
				Email:          "info@example.com",
				Logo:           "https://example.com/logo.png",
				Website:        "https://example.com",
				Status:         model.OrganizationStatusActive,
				NamespaceCount: convert.ToPointer(int64(0)),
				TeamCount:      convert.ToPointer(int64(0)),
				MemberCount:    convert.ToPointer(int64(0)),
			},
		},
		{
			name: "update organization with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Organization) []repository.RedisRepositoryOption {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateOrganizationOpts, _ *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				opts: repository.UpdateOrganizationOpts{
					Name: optional.Some("updated organization"),
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update organization set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))

					dbClient := mockrepo.NewMockUniversalClient(ctrl)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateOrganizationOpts, organization *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				opts: repository.UpdateOrganizationOpts{
					Name: optional.Some("updated organization"),
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update organization delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *repository.Organization) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(repository.OrganizationDetailProjection()))
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", organization.Slug, projectionCacheValue(repository.OrganizationDetailProjection())),
						Value: organization,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateOrganizationOpts, organization *repository.Organization) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				opts: repository.UpdateOrganizationOpts{
					Name: optional.Some("updated organization"),
				},
			},
			want: &repository.Organization{
				ID:             model.MustNewID(model.ResourceTypeOrganization),
				Slug:           "acme-org",
				Name:           "test organization",
				Email:          "info@example.com",
				Logo:           "https://example.com/logo.png",
				Website:        "https://example.com",
				Status:         model.OrganizationStatusActive,
				NamespaceCount: convert.ToPointer(int64(0)),
				TeamCount:      convert.ToPointer(int64(0)),
				MemberCount:    convert.ToPointer(int64(0)),
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

			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				require.Nil(t, got)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_AddMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository
	}
	type args struct {
		ctx      context.Context
		id       model.ID
		memberID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete organization success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().AddMember(ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "delete organization with organization deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().AddMember(ctx, id, memberID).Return(repository.ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.AddMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_RemoveMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository
	}
	type args struct {
		ctx      context.Context
		id       model.ID
		memberID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete organization success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().RemoveMember(ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "delete organization with organization deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().RemoveMember(ctx, id, memberID).Return(repository.ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.RemoveMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.OrganizationRepository
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
			name: "delete organization success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")
					refKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					refKeyCmd := new(redis.StringSliceCmd)
					refKeyCmd.SetVal([]string{refKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, refKey).Return(refKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, refKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
		},
		{
			name: "delete organization with organization deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")
					refKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetByRef", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					refKeyCmd := new(redis.StringSliceCmd)
					refKeyCmd.SetVal([]string{refKey})

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.EXPECT().Keys(ctx, refKey).Return(refKeyCmd)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, refKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_AddInvitation(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository
	}
	type args struct {
		ctx    context.Context
		orgID  model.ID
		userID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add invitation success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().AddInvitation(ctx, orgID, userID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "add invitation with organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().AddInvitation(ctx, orgID, userID).Return(repository.ErrOrganizationAddMember)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrOrganizationAddMember,
		},
		{
			name: "add invitation with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add invitation cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.AddInvitation(tt.args.ctx, tt.args.orgID, tt.args.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_RemoveInvitation(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository
	}
	type args struct {
		ctx    context.Context
		orgID  model.ID
		userID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove invitation success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "remove invitation with organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(repository.ErrOrganizationRemoveMember)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrOrganizationRemoveMember,
		},
		{
			name: "remove invitation with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "remove invitation cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", orgID.String(), "*")
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mockrepo.NewMockOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
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
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.RemoveInvitation(tt.args.ctx, tt.args.orgID, tt.args.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_GetInvitations(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) []repository.RedisRepositoryOption
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, invitations []*repository.OrganizationMember) repository.OrganizationRepository
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.OrganizationMember
		wantErr error
	}{
		{
			name: "get invitations success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) []repository.RedisRepositoryOption {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, invitations []*repository.OrganizationMember) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().GetInvitations(ctx, orgID).Return(invitations, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*repository.OrganizationMember{
				{
					ID:    model.MustNewID(model.ResourceTypeUser),
					Email: "user1@example.com",
					Roles: []string{},
				},
			},
		},
		{
			name: "get invitations with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) []repository.RedisRepositoryOption {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ []*repository.OrganizationMember) repository.OrganizationRepository {
					repo := mockrepo.NewMockOrganizationRepository(ctrl)
					repo.EXPECT().GetInvitations(ctx, orgID).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedOrganizationRepository {
				r, err := repository.NewCachedOrganizationRepository(
					tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.GetInvitations(tt.args.ctx, tt.args.orgID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
