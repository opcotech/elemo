package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedOrganizationRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, owner model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, ownerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Create", ctx, owner, organization).Return(nil)
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
				cacheRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, ownerKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Create", ctx, owner, organization).Return(repository.ErrOrganizationCreate)
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
				cacheRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) *baseRepository {
					ownerKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					ownerKeyResult := new(redis.StringSliceCmd)
					ownerKeyResult.SetVal([]string{ownerKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, ownerKey).Return(ownerKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, ownerKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, owner model.ID, organization *model.Organization) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
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
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.owner, tt.args.organization),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.owner, tt.args.organization),
			}
			err := r.Create(tt.args.ctx, tt.args.owner, tt.args.organization)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Get", ctx, id).Return(organization, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *model.Organization {
				return &model.Organization{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(organization, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *model.Organization {
				return &model.Organization{
					ID:         model.MustNewID(model.ResourceTypeOrganization),
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Get", ctx, id).Return(organization, nil)
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
			var want *model.Organization
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedOrganizationRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository
		organizationRepo func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("GetAll", ctx, offset, limit).Return(organizations, nil)
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
				cacheRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(organizations, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
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
				cacheRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("GetAll", ctx, offset, limit).Return(nil, repository.ErrNotFound)
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
			name: "get get organizations cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
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
				cacheRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organizations,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, offset, limit int, organizations []*model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("GetAll", ctx, offset, limit).Return(organizations, nil)
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
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository
		organizationRepo func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Update", ctx, id, patch).Return(organization, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepository),
						tracer: new(testMock.Tracer),
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Update", ctx, id, patch).Return(organization, nil)
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
				cacheRepo: func(ctx context.Context, id model.ID, organization *model.Organization) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(errors.New("error"))
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: organization,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID, patch map[string]any, organization *model.Organization) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Update", ctx, id, patch).Return(organization, nil)
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

			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedOrganizationRepository_AddMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, id, memberID model.ID) *baseRepository
		organizationRepo func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("AddMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete organization with organization deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("AddMember", ctx, id, memberID).Return(repository.ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("AddMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.AddMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_RemoveMember(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, id, memberID model.ID) *baseRepository
		organizationRepo func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete organization with organization deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(repository.ErrOrganizationDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrOrganizationDelete,
		},
		{
			name: "delete organization with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("RemoveMember", ctx, id, memberID).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete organization cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, memberID model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id, memberID model.ID) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeOrganization),
				memberID: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.id, tt.args.memberID),
			}
			err := r.RemoveMember(tt.args.ctx, tt.args.id, tt.args.memberID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedOrganizationRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo        func(ctx context.Context, id model.ID) *baseRepository
		organizationRepo func(ctx context.Context, id model.ID) repository.OrganizationRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrOrganizationDelete)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())

					dbClient := new(testMock.RedisClient)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID) repository.OrganizationRepository {
					repo := new(testMock.OrganizationRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				organizationRepo: func(ctx context.Context, id model.ID) repository.OrganizationRepository {
					return new(testMock.OrganizationRepository)
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
			r := &CachedOrganizationRepository{
				cacheRepo:        tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				organizationRepo: tt.fields.organizationRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
