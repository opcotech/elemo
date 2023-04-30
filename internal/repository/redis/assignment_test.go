package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedAssignmentRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, assignment *model.Assignment) *baseRepository
		assignmentRepo func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository
	}
	type args struct {
		ctx        context.Context
		assignment *model.Assignment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create new document assignment",
			fields: fields{
				cacheRepo: func(ctx context.Context, assignment *model.Assignment) *baseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String())
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String())
					key3 := composeCacheKey(model.ResourceTypeDocument.String(), assignment.Resource.String())

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")
					resourceKey := composeCacheKey(model.ResourceTypeDocument.String(), "*")

					byResourceKeyResult := new(redis.StringSliceCmd)
					byResourceKeyResult.SetVal([]string{key1})

					byUserKeyResult := new(redis.StringSliceCmd)
					byUserKeyResult.SetVal([]string{key2})

					resourceKeyResult := new(redis.StringSliceCmd)
					resourceKeyResult.SetVal([]string{key3})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byResourceKey).Return(byResourceKeyResult)
					dbClient.On("Keys", ctx, byUserKey).Return(byUserKeyResult)
					dbClient.On("Keys", ctx, resourceKey).Return(resourceKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key1).Return(nil)
					cacheRepo.On("Delete", ctx, key2).Return(nil)
					cacheRepo.On("Delete", ctx, key3).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Create", ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeDocument),
				},
			},
		},
		{
			name: "create new issue assignment",
			fields: fields{
				cacheRepo: func(ctx context.Context, assignment *model.Assignment) *baseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String())
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String())
					key3 := composeCacheKey(model.ResourceTypeIssue.String(), assignment.Resource.String())

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")
					resourceKey := composeCacheKey(model.ResourceTypeIssue.String(), "*")

					byResourceKeyResult := new(redis.StringSliceCmd)
					byResourceKeyResult.SetVal([]string{key1})

					byUserKeyResult := new(redis.StringSliceCmd)
					byUserKeyResult.SetVal([]string{key2})

					resourceKeyResult := new(redis.StringSliceCmd)
					resourceKeyResult.SetVal([]string{key3})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byResourceKey).Return(byResourceKeyResult)
					dbClient.On("Keys", ctx, byUserKey).Return(byUserKeyResult)
					dbClient.On("Keys", ctx, resourceKey).Return(resourceKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key1).Return(nil)
					cacheRepo.On("Delete", ctx, key2).Return(nil)
					cacheRepo.On("Delete", ctx, key3).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Create", ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeIssue),
				},
			},
		},
		{
			name: "create new unknown resource assignment",
			fields: fields{
				cacheRepo: func(ctx context.Context, assignment *model.Assignment) *baseRepository {
					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.On("Keys", ctx, byUserKey).Return(new(redis.StringSliceCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseRepository{
						db:     db,
						cache:  new(testMock.CacheRepo),
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Create", ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: ErrUnexpectedCachedResource,
		},
		{
			name: "create new assignment with by resource cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, assignment *model.Assignment) *baseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byResourceKey).Return(keysCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key1).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Create", ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "create new assignment with by user cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, assignment *model.Assignment) *baseRepository {
					key1 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.Resource.String(), "1")
					key2 := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.Resource.String(), "2")

					byResourceKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", assignment.Resource.String(), "*")
					byUserKey := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", assignment.User.String(), "*")

					keysCmd := new(redis.StringSliceCmd)
					keysCmd.SetVal([]string{key1, key2})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, byResourceKey).Return(new(redis.StringSliceCmd))
					dbClient.On("Keys", ctx, byUserKey).Return(keysCmd)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Delete", ctx, key1).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Create", ctx, assignment).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				assignment: &model.Assignment{
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.assignment),
				assignmentRepo: tt.fields.assignmentRepo(tt.args.ctx, tt.args.assignment),
			}
			err := r.Create(tt.args.ctx, tt.args.assignment)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedAssignmentRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository
		assignmentRepo func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Assignment
		wantErr error
	}{
		{
			name: "get uncached assignment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: assignment,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Get", ctx, id).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *model.Assignment {
				return &model.Assignment{
					ID:       id,
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				}
			},
		},
		{
			name: "get cached assignment",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(assignment, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository {
					return new(testMock.AssignmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			want: func(id model.ID) *model.Assignment {
				return &model.Assignment{
					ID:       id,
					Kind:     model.AssignmentKindAssignee,
					User:     model.MustNewID(model.ResourceTypeUser),
					Resource: model.MustNewID(model.ResourceTypeProject),
				}
			},
		},
		{
			name: "get uncached assignment error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached assignment error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository {
					return new(testMock.AssignmentRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached assignment cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) *baseRepository {
					key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepo)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: assignment,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				assignmentRepo: func(ctx context.Context, id model.ID, assignment *model.Assignment) repository.AssignmentRepository {
					repo := new(testMock.AssignmentRepository)
					repo.On("Get", ctx, id).Return(assignment, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeAssignment),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var want *model.Assignment
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				assignmentRepo: tt.fields.assignmentRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, want, got)
		})
	}
}

func TestCachedAssignmentRepository_GetByResource(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, resourceID model.ID, offset, limit int) *baseRepository
		assignmentRepo func(ctx context.Context, resourceID model.ID, offset, limit int) repository.AssignmentRepository
	}
	type args struct {
		ctx        context.Context
		resourceID model.ID
		offset     int
		limit      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Assignment
		wantErr error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.resourceID, tt.args.offset, tt.args.limit),
				assignmentRepo: tt.fields.assignmentRepo(tt.args.ctx, tt.args.resourceID, tt.args.offset, tt.args.limit),
			}
			got, err := r.GetByResource(tt.args.ctx, tt.args.resourceID, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAssignmentRepository_GetByUser(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, userID model.ID, offset, limit int) *baseRepository
		assignmentRepo func(ctx context.Context, userID model.ID, offset, limit int) repository.AssignmentRepository
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
		want    []*model.Assignment
		wantErr error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit),
				assignmentRepo: tt.fields.assignmentRepo(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit),
			}
			got, err := r.GetByUser(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.limit)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedAssignmentRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo      func(ctx context.Context, id model.ID) *baseRepository
		assignmentRepo func(ctx context.Context, id model.ID) repository.AssignmentRepository
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CachedAssignmentRepository{
				cacheRepo:      tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				assignmentRepo: tt.fields.assignmentRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
