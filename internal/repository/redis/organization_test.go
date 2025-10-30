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

func TestCachedOrganizationRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Create(ctx, owner, organization).Return(repository.ErrOrganizationCreate)
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
			wantErr: repository.ErrOrganizationCreate,
		},
		{
			name: "add new organization get all cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, ownerKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) repository.OrganizationRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &CachedOrganizationRepository{
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
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						if orgPtr, ok := dst.(**model.Organization); ok {
							*orgPtr = organization
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
						Value: organization,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Get(ctx, id).Return(organization, nil)
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
			var want *model.Organization
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedOrganizationRepository{
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
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository
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
		want    []*model.Organization
		wantErr error
	}{
		{
			name: "get uncached organizations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(organizations, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

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
						if orgsPtr, ok := dst.(*[]*model.Organization); ok {
							*orgsPtr = organizations
						}
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*model.Organization) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get organizations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, _ []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ int, _ []*model.Organization) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached organizations cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

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
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().GetAll(ctx, offset, limit).Return(organizations, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
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
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *model.Organization) *baseRepository {
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
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, _ *model.Organization) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
					repo.EXPECT().Update(ctx, id, patch).Return(nil, repository.ErrNotFound)
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
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update organization set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := mock.NewUniversalClient(ctrl)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update organization delete get all cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redis.baseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &CachedOrganizationRepository{
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
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) *baseRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
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
			r := &CachedOrganizationRepository{
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
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) *baseRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
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
			r := &CachedOrganizationRepository{
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
		cacheRepo        func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.OrganizationRepository {
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
					cacheRepo.EXPECT().Delete(ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				organizationRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.OrganizationRepository {
					repo := mock.NewOrganizationRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

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
				organizationRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) repository.OrganizationRepository {
					return mock.NewOrganizationRepository(ctrl)
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
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				organizationRepo: tt.fields.organizationRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
