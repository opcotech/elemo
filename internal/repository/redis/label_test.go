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

func TestCachedLabelRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, label *model.Label) *baseRepository
		labelRepo func(ctx context.Context, label *model.Label) repository.LabelRepository
	}
	type args struct {
		ctx   context.Context
		label *model.Label
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new label",
			fields: fields{
				cacheRepo: func(ctx context.Context, label *model.Label) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Create", ctx, label).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				label: &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
		},
		{
			name: "add new label with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, label *model.Label) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Create", ctx, label).Return(repository.ErrLabelCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				label: &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
			wantErr: repository.ErrLabelCreate,
		},
		{
			name: "add new label get all cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, label *model.Label) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, getAllKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, label *model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				label: &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new label documents cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, label *model.Label) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, label *model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				label: &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new label issues cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, label *model.Label) *baseRepository {
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyResult := new(redis.StringSliceCmd)
					getAllKeyResult.SetVal([]string{getAllKey})

					documentsKeyResult := new(redis.StringSliceCmd)
					documentsKeyResult.SetVal([]string{documentsKey})

					issuesKeyResult := new(redis.StringSliceCmd)
					issuesKeyResult.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyResult)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyResult)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, getAllKey).Return(nil)
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, label *model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				label: &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.label),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.label),
			}
			err := r.Create(tt.args.ctx, tt.args.label)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedLabelRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, label *model.Label) *baseRepository
		labelRepo func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Label
		wantErr error
	}{
		{
			name: "get uncached label",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
						Value: label,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Get", ctx, id).Return(label, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			want: func(id model.ID) *model.Label {
				return &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				}
			},
		},
		{
			name: "get cached label",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(label, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			want: func(id model.ID) *model.Label {
				return &model.Label{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				}
			},
		},
		{
			name: "get uncached label error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
				labelRepo: func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached label error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
				labelRepo: func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached label cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
						Value: label,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Get", ctx, id).Return(label, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want *model.Label
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedLabelRepository_GetAll(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository
		labelRepo func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository
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
		want    []*model.Label
		wantErr error
	}{
		{
			name: "get uncached labels",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)

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
						Value: labels,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("GetAll", ctx, offset, limit).Return(labels, nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*model.Label{
				{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
				{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
		},
		{
			name: "get cached labels",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(labels, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want: []*model.Label{
				{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
				{
					ID:          model.MustNewID(model.ResourceTypeLabel),
					Name:        "test label",
					Description: "test description",
				},
			},
		},
		{
			name: "get uncached labels error",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)

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
				labelRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
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
			name: "get get labels cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)

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
				labelRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository {
					return new(testMock.LabelRepository)
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
			name: "get uncached labels cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", offset, limit)

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
						Value: labels,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, offset, limit int, labels []*model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("GetAll", ctx, offset, limit).Return(labels, nil)
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
			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedLabelRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, label *model.Label) *baseRepository
		labelRepo func(ctx context.Context, id model.ID, patch map[string]any, label *model.Label) repository.LabelRepository
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
		want    *model.Label
		wantErr error
	}{
		{
			name: "update label",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: label,
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
						Value: label,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, patch map[string]any, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Update", ctx, id, patch).Return(label, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
				patch: map[string]any{
					"name":        "updated label",
					"description": "updated description",
				},
			},
			want: &model.Label{
				ID:          model.MustNewID(model.ResourceTypeLabel),
				Name:        "test label",
				Description: "test description",
			},
		},
		{
			name: "update label with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
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
				labelRepo: func(ctx context.Context, id model.ID, patch map[string]any, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
				patch: map[string]any{
					"name":        "updated label",
					"description": "updated description",
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update label set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: label,
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
						Value: label,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, patch map[string]any, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Update", ctx, id, patch).Return(label, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
				patch: map[string]any{
					"name":        "updated label",
					"description": "updated description",
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update label delete get all cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, label *model.Label) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: label,
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
						Value: label,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID, patch map[string]any, label *model.Label) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Update", ctx, id, patch).Return(label, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
				patch: map[string]any{
					"name":        "updated label",
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

			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedLabelRepository_AttachTo(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id, attachTo model.ID) *baseRepository
		labelRepo func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository
	}
	type args struct {
		ctx      context.Context
		id       model.ID
		attachTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete label success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("AttachTo", ctx, id, attachTo).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete label with label deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("AttachTo", ctx, id, attachTo).Return(repository.ErrLabelDelete)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrLabelDelete,
		},
		{
			name: "delete label with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("AttachTo", ctx, id, attachTo).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")

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
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by document key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by issues key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, attachTo model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, attachTo model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:      context.Background(),
				id:       model.MustNewID(model.ResourceTypeLabel),
				attachTo: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.attachTo),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.id, tt.args.attachTo),
			}
			err := r.AttachTo(tt.args.ctx, tt.args.id, tt.args.attachTo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedLabelRepository_DetachFrom(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id, detachFrom model.ID) *baseRepository
		labelRepo func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository
	}
	type args struct {
		ctx        context.Context
		id         model.ID
		detachFrom model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete label success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("DetachFrom", ctx, id, detachFrom).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
		},
		{
			name: "delete label with label deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("DetachFrom", ctx, id, detachFrom).Return(repository.ErrLabelDelete)
					return repo
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrLabelDelete,
		},
		{
			name: "delete label with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("DetachFrom", ctx, id, detachFrom).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")

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
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by document key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by issues key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, detachFrom model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id, detachFrom model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx:        context.Background(),
				id:         model.MustNewID(model.ResourceTypeLabel),
				detachFrom: model.MustNewID(model.ResourceTypeDocument),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.detachFrom),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.id, tt.args.detachFrom),
			}
			err := r.DetachFrom(tt.args.ctx, tt.args.id, tt.args.detachFrom)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedLabelRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID) *baseRepository
		labelRepo func(ctx context.Context, id model.ID) repository.LabelRepository
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
			name: "delete label success",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
		},
		{
			name: "delete label with label deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrLabelDelete)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrLabelDelete,
		},
		{
			name: "delete label with cache deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())

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
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					repo := new(testMock.LabelRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by related key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")

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
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by document key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete label cache by issues key error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeLabel.String(), id.String())
					getAllKey := composeCacheKey(model.ResourceTypeLabel.String(), "GetAll", "*")
					documentsKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")
					issuesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					getAllKeyCmd := new(redis.StringSliceCmd)
					getAllKeyCmd.SetVal([]string{getAllKey})

					documentsKeyCmd := new(redis.StringSliceCmd)
					documentsKeyCmd.SetVal([]string{documentsKey})

					issuesKeyCmd := new(redis.StringSliceCmd)
					issuesKeyCmd.SetVal([]string{issuesKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, getAllKey).Return(getAllKeyCmd)
					dbClient.On("Keys", ctx, documentsKey).Return(documentsKeyCmd)
					dbClient.On("Keys", ctx, issuesKey).Return(issuesKeyCmd)

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
					cacheRepo.On("Delete", ctx, documentsKey).Return(nil)
					cacheRepo.On("Delete", ctx, issuesKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				labelRepo: func(ctx context.Context, id model.ID) repository.LabelRepository {
					return new(testMock.LabelRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeLabel),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedLabelRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				labelRepo: tt.fields.labelRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
