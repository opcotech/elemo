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

func TestCachedOrganizationRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) OrganizationRepository
	}
	type args struct {
		ctx          context.Context
		owner        model.ID
		organization *model.Organization
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *redisBaseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Create(ctx, owner, organization).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				owner: model.MustNewID(model.ResourceTypeUser),
				organization: &model.Organization{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
			},
		},
		{
			name: "add new organization with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *redisBaseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Create(ctx, owner, organization).Return(ErrOrganizationCreate)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				owner: model.MustNewID(model.ResourceTypeUser),
				organization: &model.Organization{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
			},
			wantErr: ErrOrganizationCreate,
		},
		{
			name: "add new organization get all cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *redisBaseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				owner: model.MustNewID(model.ResourceTypeUser),
				organization: &model.Organization{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.organization),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.owner, tt.args.organization),
			}
			err := r.Create(tt.args.ctx, tt.args.owner, tt.args.organization)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) OrganizationRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Organization
		wantErr error
	}{
		{
			name: "get uncached organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						Value: organization,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *model.Organization {
				return &model.Organization{
					ID:         id,
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				}
			},
		},
		{
			name: "get cached organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						if orgPtr, ok := dst.(**model.Organization); ok {
							*orgPtr = organization
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *model.Organization {
				return &model.Organization{
					ID:         id,
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				}
			},
		},
		{
			name: "get uncached organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached organization error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached organization cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						Value: organization,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
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
			var want *model.Organization
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedOrganizationRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) OrganizationRepository
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
		want    []*model.Organization
		wantErr error
	}{
		{
			name: "get uncached organizations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, userID, offset, limit).Return(organizations, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*model.Organization{
				{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
				{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
			},
		},
		{
			name: "get cached organizations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)

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
						if orgsPtr, ok := dst.(*[]*model.Organization); ok {
							*orgsPtr = organizations
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Organization) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			want: []*model.Organization{
				{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
				{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
					Name:       "test organization",
					Email:      "info@example.com",
					Logo:       "https://example.com/logo.png",
					Website:    "https://example.com",
					Status:     model.OrganizationStatusActive,
					Namespaces: make([]model.ID, 0),
					Teams:      make([]model.ID, 0),
					Members:    make([]model.ID, 0),
				},
			},
		},
		{
			name: "get uncached organizations error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, userID, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get organizations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, _ []*model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*model.Organization) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached organizations cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID, offset, limit int, organizations []*model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, userID, offset, limit).Return(organizations, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: model.MustNewID(model.ResourceTypeUser),
				offset: 0,
				limit:  10,
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) OrganizationRepository
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
		want    *model.Organization
		wantErr error
	}{
		{
			name: "update organization",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				patch: map[string]any{
					"name":        "updated organization",
					"description": "updated description",
				},
			},
			want: &model.Organization{
				ID:         model.MustNewID(model.ResourceTypeOrganization),
				Name:       "test organization",
				Email:      "info@example.com",
				Logo:       "https://example.com/logo.png",
				Website:    "https://example.com",
				Status:     model.OrganizationStatusActive,
				Namespaces: make([]model.ID, 0),
				Teams:      make([]model.ID, 0),
				Members:    make([]model.ID, 0),
			},
		},
		{
			name: "update organization with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) *redisBaseRepository {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				patch: map[string]any{
					"name":        "updated organization",
					"description": "updated description",
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update organization set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				patch: map[string]any{
					"name":        "updated organization",
					"description": "updated description",
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update organization delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
				patch: map[string]any{
					"name":        "updated organization",
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

			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_AddMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().AddMember(ctx, id, memberID).Return(ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.AddMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_RemoveMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().RemoveMember(ctx, id, memberID).Return(ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.RemoveMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Delete(ctx, id).Return(ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_AddInvitation(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().AddInvitation(ctx, orgID, userID).Return(ErrOrganizationAddMember)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrOrganizationAddMember,
		},
		{
			name: "add invitation with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add invitation cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
			}
			err := r.AddInvitation(tt.args.ctx, tt.args.orgID, tt.args.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_RemoveInvitation(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, userID model.ID) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().RemoveInvitation(ctx, orgID, userID).Return(ErrOrganizationRemoveMember)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrOrganizationRemoveMember,
		},
		{
			name: "remove invitation with cache deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "remove invitation cache by related key error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), orgID.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				orgID:  model.MustNewID(model.ResourceTypeOrganization),
				userID: model.MustNewID(model.ResourceTypeUser),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.args.userID),
			}
			err := r.RemoveInvitation(tt.args.ctx, tt.args.orgID, tt.args.userID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_GetInvitations(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID) *redisBaseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, invitations []*model.OrganizationMember) OrganizationRepository
	}
	type args struct {
		ctx   context.Context
		orgID model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.OrganizationMember
		wantErr error
	}{
		{
			name: "get invitations success",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) *redisBaseRepository {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, invitations []*model.OrganizationMember) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetInvitations(ctx, orgID).Return(invitations, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: []*model.OrganizationMember{
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) *redisBaseRepository {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, orgID model.ID, _ []*model.OrganizationMember) OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetInvitations(ctx, orgID).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				orgID: model.MustNewID(model.ResourceTypeOrganization),
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
			r := &RedisCachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.orgID),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.orgID, tt.want),
			}
			got, err := r.GetInvitations(tt.args.ctx, tt.args.orgID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}
