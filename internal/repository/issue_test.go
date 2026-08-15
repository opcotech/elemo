package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func issueRelationPairCachePatterns(source, target model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeIssue.String(), "Get", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "Get", target.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", target.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", target.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*"),
	}
}

func TestCachedIssueRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) IssueRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{{
		name: "add new issue with no parent",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")
				namespacesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForNamespace", "*")

				allProjectsKeyResult := new(redis.StringSliceCmd)
				allProjectsKeyResult.SetVal([]string{allProjectsKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				namespacesKeyResult := new(redis.StringSliceCmd)
				namespacesKeyResult.SetVal([]string{namespacesKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
				dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(4)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				repo.EXPECT().Create(ctx, opts).Return(&Issue{}, nil)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      nil,
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
	}, {
		name: "add new issue with parent",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")
				namespacesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForNamespace", "*")
				parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", opts.Parent.String(), "*")

				allProjectsKeyResult := new(redis.StringSliceCmd)
				allProjectsKeyResult.SetVal([]string{allProjectsKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				namespacesKeyResult := new(redis.StringSliceCmd)
				namespacesKeyResult.SetVal([]string{namespacesKey})

				parentIssueKeyResult := new(redis.StringSliceCmd)
				parentIssueKeyResult.SetVal([]string{parentIssueKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
				dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
				dbClient.EXPECT().Keys(ctx, parentIssueKey).Return(parentIssueKeyResult)

				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(5)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, parentIssueKey).Return(nil)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				repo.EXPECT().Create(ctx, opts).Return(&Issue{}, nil)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
	}, {
		name: "add new issue with error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")
				namespacesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForNamespace", "*")

				allProjectsKeyResult := new(redis.StringSliceCmd)
				allProjectsKeyResult.SetVal([]string{allProjectsKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				namespacesKeyResult := new(redis.StringSliceCmd)
				namespacesKeyResult.SetVal([]string{namespacesKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
				dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)

				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(4)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				repo.EXPECT().Create(ctx, opts).Return(nil, ErrIssueCreate)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      nil,
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
		wantErr: ErrIssueCreate,
	}, {
		name: "add new issue with cache delete error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")
				namespacesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForNamespace", "*")

				allProjectsKeyResult := new(redis.StringSliceCmd)
				allProjectsKeyResult.SetVal([]string{allProjectsKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				namespacesKeyResult := new(redis.StringSliceCmd)
				namespacesKeyResult.SetVal([]string{namespacesKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
				dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
				dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(3)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueOpts) IssueRepository {
				return NewMockIssueRepository(ctrl)
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      nil,
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "add new issue with parent issue cache delete error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")
				namespacesKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForNamespace", "*")
				parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", opts.Parent.String(), "*")

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				namespacesKeyResult := new(redis.StringSliceCmd)
				namespacesKeyResult.SetVal([]string{namespacesKey})

				parentIssueKeyResult := new(redis.StringSliceCmd)
				parentIssueKeyResult.SetVal([]string{parentIssueKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
				dbClient.EXPECT().Keys(ctx, namespacesKey).Return(namespacesKeyResult)
				dbClient.EXPECT().Keys(ctx, parentIssueKey).Return(parentIssueKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(3)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, namespacesKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, parentIssueKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueOpts) IssueRepository {
				return NewMockIssueRepository(ctrl)
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "add new issue with project cache delete error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
				projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForProject", opts.ProjectID.String(), "*")

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(1)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueOpts) IssueRepository {
				return NewMockIssueRepository(ctrl)
			},
		},
		args: args{
			ctx: context.Background(),
			opts: CreateIssueOpts{
				ProjectID:   model.MustNewID(model.ResourceTypeProject),
				Parent:      nil,
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Links:       make([]model.IssueLink, 0),
			},
		},
		wantErr: ErrCacheDelete,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := r.Create(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *Issue
		wantErr error
	}{
		{
			name: "get uncached issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
						Value: issue,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *Issue {
				return &Issue{
					ID:              id,
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				}
			},
		},
		{
			name: "get cached issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
						if ptr, ok := dst.(**Issue); ok {
							*ptr = issue
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Issue) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *Issue {
				return &Issue{
					ID:              id,
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				}
			},
		},
		{
			name: "get uncached issue error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, IssueDetailProjection()).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached issue error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Issue) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached issue cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
						Value: issue,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var want *Issue
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id, IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedIssueRepository_GetByKey(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *Issue) IssueRepository
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		issueKey    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Issue
		wantErr error
	}{
		{
			name: "get uncached issue by key",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", namespaceID.String(), issueKey, projectionCacheValue(IssueDetailProjection()))

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
						Value: issue,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, namespaceID, issueKey, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				issueKey:    "ENG-42",
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				Key:       "ENG-42",
				NumericID: 42,
				Kind:      model.IssueKindStory,
				Title:     "test issue",
			},
		},
		{
			name: "get cached issue by key",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", namespaceID.String(), issueKey, projectionCacheValue(IssueDetailProjection()))

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
						if ptr, ok := dst.(**Issue); ok {
							*ptr = issue
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ string, _ *Issue) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				issueKey:    "ENG-42",
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				Key:       "ENG-42",
				NumericID: 42,
				Kind:      model.IssueKindStory,
				Title:     "test issue",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, tt.want),
			}
			got, err := r.GetByKey(tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForProject(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query IssueListQuery, page Page[*PartialIssue]) IssueRepository
	}
	type args struct {
		ctx   context.Context
		query IssueListQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						Value: page,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query IssueListQuery, page Page[*PartialIssue]) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForProject(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListQuery{
					ProjectID:  model.MustNewID(model.ResourceTypeProject),
					Page:       CursorPage{Size: 10},
					Projection: IssueListForProjectProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "PROJ-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						*(dst.(*Page[*PartialIssue])) = page
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ IssueListQuery, _ Page[*PartialIssue]) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListQuery{
					ProjectID:  model.MustNewID(model.ResourceTypeProject),
					Page:       CursorPage{Size: 10},
					Projection: IssueListForProjectProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "PROJ-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			key, err := issueListForProjectCacheKey(tt.args.query)
			require.NoError(t, err)

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, key, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
			}
			got, err := r.ListForProject(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForNamespace(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query IssueListForNamespaceQuery, page Page[*PartialIssue]) IssueRepository
	}
	type args struct {
		ctx   context.Context
		query IssueListForNamespaceQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						Value: page,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query IssueListForNamespaceQuery, page Page[*PartialIssue]) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForNamespace(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListForNamespaceQuery{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					Page:        CursorPage{Size: 10},
					Projection:  IssueListForNamespaceProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						*(dst.(*Page[*PartialIssue])) = page
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ IssueListForNamespaceQuery, _ Page[*PartialIssue]) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListForNamespaceQuery{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					Page:        CursorPage{Size: 10},
					Projection:  IssueListForNamespaceProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			key, err := issueListForNamespaceCacheKey(tt.args.query)
			require.NoError(t, err)

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, key, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
			}
			got, err := r.ListForNamespace(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForUser(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query IssueListForUserQuery, page Page[*PartialIssue]) IssueRepository
	}
	type args struct {
		ctx   context.Context
		query IssueListForUserQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page[*PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						Value: page,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query IssueListForUserQuery, page Page[*PartialIssue]) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForUser(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListForUserQuery{
					UserID:     model.MustNewID(model.ResourceTypeUser),
					Page:       CursorPage{Size: 10},
					Projection: IssueListForUserProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, key string, page Page[*PartialIssue]) *redisBaseRepository {
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
						*(dst.(*Page[*PartialIssue])) = page
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ IssueListForUserQuery, _ Page[*PartialIssue]) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: IssueListForUserQuery{
					UserID:     model.MustNewID(model.ResourceTypeUser),
					Page:       CursorPage{Size: 10},
					Projection: IssueListForUserProjection(),
				},
			},
			want: Page[*PartialIssue]{
				Items: []*PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]PartialAssignee, 0),
						Labels:      make([]PartialLabel, 0),
					},
				},
				PageInfo: PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			key, err := issueListForUserCacheKey(tt.args.query)
			require.NoError(t, err)

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, key, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
			}
			got, err := r.ListForUser(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForIssue(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) IssueRepository
	}
	type args struct {
		ctx    context.Context
		issue  model.ID
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Issue
		wantErr error
	}{
		{
			name: "get uncached issues",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) *redisBaseRepository {
					key := issueListForIssueCacheKey(issue, CursorPage{Size: testPageSize(limit)}, IssueDetailProjection())

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
						Value: Page[*Issue]{Items: issues},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, IssueListForIssueQuery{IssueID: issue, Page: CursorPage{Size: limit}, Projection: IssueDetailProjection()}).Return(Page[*Issue]{Items: issues}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*Issue{
				{
					ID:              model.MustNewID(model.ResourceTypeIssue),
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeIssue),
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				},
			},
		},
		{
			name: "get cached issues",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) *redisBaseRepository {
					key := issueListForIssueCacheKey(issue, CursorPage{Size: testPageSize(limit)}, IssueDetailProjection())

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
						if ptr, ok := dst.(*Page[*Issue]); ok {
							*ptr = Page[*Issue]{Items: issues}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Issue) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*Issue{
				{
					ID:              model.MustNewID(model.ResourceTypeIssue),
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				},
				{
					ID:              model.MustNewID(model.ResourceTypeIssue),
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]PartialAssignee, 0),
					Labels:          make([]PartialLabel, 0),
					CommentCount:    convert.ToPointer(int64(0)),
					AttachmentCount: convert.ToPointer(int64(0)),
					WatcherCount:    convert.ToPointer(int64(0)),
					RelationCount:   convert.ToPointer(int64(0)),
					Links:           make([]model.IssueLink, 0),
				},
			},
		},
		{
			name: "get uncached issues error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*Issue) *redisBaseRepository {
					key := issueListForIssueCacheKey(issue, CursorPage{Size: testPageSize(limit)}, IssueDetailProjection())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, IssueListForIssueQuery{IssueID: issue, Page: CursorPage{Size: limit}, Projection: IssueDetailProjection()}).Return(Page[*Issue]{}, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get issues cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*Issue) *redisBaseRepository {
					key := issueListForIssueCacheKey(issue, CursorPage{Size: testPageSize(limit)}, IssueDetailProjection())

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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*Issue) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached issues cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) *redisBaseRepository {
					key := issueListForIssueCacheKey(issue, CursorPage{Size: testPageSize(limit)}, IssueDetailProjection())

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
						Value: Page[*Issue]{Items: issues},
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, IssueListForIssueQuery{IssueID: issue, Page: CursorPage{Size: limit}, Projection: IssueDetailProjection()}).Return(Page[*Issue]{Items: issues}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, testPageSize(tt.args.limit), tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, testPageSize(tt.args.limit), tt.want),
			}
			got, err := r.ListForIssue(tt.args.ctx, IssueListForIssueQuery{IssueID: tt.args.issue, Page: CursorPage{Size: testPageSize(tt.args.limit)}, Projection: IssueDetailProjection()})
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedIssueRepository_AddWatcher(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository
	}
	type args struct {
		ctx     context.Context
		id      model.ID
		watcher model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "add watcher",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().AddWatcher(ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
		},
		{
			name: "add watcher with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().AddWatcher(ctx, id, watcher).Return(ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "add watcher with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					return redisCacheExpectingPatterns(ctrl, ctx, []string{key}, 0, ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add watcher with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},

		{
			name: "add watcher with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add watcher with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
			}
			err := r.AddWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetWatchers(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*User
		wantErr error
	}{
		{
			name: "get issue watchers",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						Value: watchers,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(watchers, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*User{
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
			},
		},
		{
			name: "get issue watchers with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*User) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get issue watchers from cache",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						if ptr, ok := dst.(*[]*User); ok {
							*ptr = watchers
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*User) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*User{
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
					Status:   model.UserStatusActive,
				},
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
					Status:   model.UserStatusActive,
				},
			},
		},
		{
			name: "get issue watchers with cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						Value: watchers,
					}).Return(ErrCacheWrite)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*User) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(watchers, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*User{
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "get issue watchers with get cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*User) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(ErrCacheRead)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*User) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*User{
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
				{
					ID:       model.MustNewID(model.ResourceTypeUser),
					Username: "test-user",
					Email:    "test@example.com",
				},
			},
			wantErr: ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetWatchers(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedIssueRepository_RemoveWatcher(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository
	}
	type args struct {
		ctx     context.Context
		id      model.ID
		watcher model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove issue watcher",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveWatcher(ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
		},
		{
			name: "remove issue watcher with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveWatcher(ctx, id, watcher).Return(ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "remove issue watcher with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					return redisCacheExpectingPatterns(ctrl, ctx, []string{key}, 0, ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					keyCmd := new(redis.StringSliceCmd)
					keyCmd.SetVal([]string{key})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

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
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
			}
			err := r.RemoveWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_AddRelation(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) IssueRepository
	}
	type args struct {
		ctx  context.Context
		opts CreateIssueRelationOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{

			name: "add issue relation",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().AddRelation(ctx, opts).Return(&IssueRelation{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "add issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().AddRelation(ctx, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: ErrNotFound,
		},
		{
			name: "add issue relation with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), 0, ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueRelationOpts) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.opts),
			}
			_, err := r.AddRelation(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetRelation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)
	want := &IssueRelation{ID: relationID, Kind: model.IssueRelationKindBlocks}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockIssueRepository(ctrl)
	repo.EXPECT().GetRelation(ctx, relationID).Return(want, nil)

	r := &RedisCachedIssueRepository{
		cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, nil, -1, nil),
		issueRepo: repo,
	}

	got, err := r.GetRelation(ctx, relationID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCachedIssueRepository_ListRelations(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	query := IssueRelationListQuery{IssueID: issueID, Page: CursorPage{Size: 10}}
	key, err := issueListRelationsCacheKey(query)
	require.NoError(t, err)
	page := Page[*IssueRelationItem]{Items: []*IssueRelationItem{{ID: model.MustNewID(model.ResourceTypeIssueRelation)}}}

	t.Run("get uncached relation page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Get", gomock.Len(0)).Return(context.Background(), span)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Set", gomock.Len(0)).Return(context.Background(), span)

		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(gomock.Any(), key, gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Set(&cache.Item{
			Ctx:   context.Background(),
			Key:   key,
			Value: page,
		}).Return(nil)

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)

		issueRepo := NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().ListRelations(gomock.Any(), query).Return(page, nil)

		r := &RedisCachedIssueRepository{
			cacheRepo: &redisBaseRepository{db: db, cache: cacheRepo, tracer: tracer, logger: mock.NewMockLogger(ctrl)},
			issueRepo: issueRepo,
		}
		got, err := r.ListRelations(context.Background(), query)
		require.NoError(t, err)
		assert.Equal(t, page, got)
	})

	t.Run("get cached relation page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Get", gomock.Len(0)).Return(context.Background(), span)

		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(gomock.Any(), key, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest any) error {
			*(dest.(*Page[*IssueRelationItem])) = page
			return nil
		})

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)

		r := &RedisCachedIssueRepository{
			cacheRepo: &redisBaseRepository{db: db, cache: cacheRepo, tracer: tracer, logger: mock.NewMockLogger(ctrl)},
			issueRepo: NewMockIssueRepository(ctrl),
		}
		got, err := r.ListRelations(context.Background(), query)
		require.NoError(t, err)
		assert.Equal(t, page, got)
	})
}

func TestCachedIssueRepository_GetRelations(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*IssueRelation
		wantErr error
	}{
		{
			name: "get issue relations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						Value: relations,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(relations, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*IssueRelation{
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "get issue relations with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*IssueRelation) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*IssueRelation) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get issue relations from cache",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						if ptr, ok := dst.(*[]*IssueRelation); ok {
							*ptr = relations
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*IssueRelation) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*IssueRelation{
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "get issue relations with cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						Value: relations,
					}).Return(ErrCacheWrite)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*IssueRelation) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(relations, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*IssueRelation{
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "get issue relations with get cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*IssueRelation) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(ErrCacheRead)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*IssueRelation) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*IssueRelation{
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
				{
					ID:     model.MustNewID(model.ResourceTypeIssueRelation),
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
			}
			got, err := r.GetRelations(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedIssueRepository_RemoveRelation(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) IssueRepository
	}
	type args struct {
		ctx    context.Context
		source model.ID
		target model.ID
		kind   model.IssueRelationKind
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "remove issue relation",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveRelation(ctx, source, target, kind).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeIssue),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
		},
		{
			name: "remove issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveRelation(ctx, source, target, kind).Return(ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeIssue),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
			wantErr: ErrNotFound,
		},
		{
			name: "remove issue relation with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), 0, ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) IssueRepository {
					return NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeIssue),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind),
			}
			err := r.RemoveRelation(tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_RemoveRelationByID(t *testing.T) {
	ctx := context.Background()
	source := model.MustNewID(model.ResourceTypeIssue)
	target := model.MustNewID(model.ResourceTypeIssue)
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)
	rel := &IssueRelation{ID: relationID, Source: source, Target: target, Kind: model.IssueRelationKindBlocks}

	t.Run("remove relation by id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		issueRepo := NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(ctx, relationID).Return(rel, nil)
		issueRepo.EXPECT().RemoveRelationByID(ctx, relationID).Return(nil)

		r := &RedisCachedIssueRepository{
			cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil),
			issueRepo: issueRepo,
		}
		require.NoError(t, r.RemoveRelationByID(ctx, relationID))
	})

	t.Run("get relation error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		issueRepo := NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(ctx, relationID).Return(nil, ErrNotFound)

		r := &RedisCachedIssueRepository{
			cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, nil, -1, nil),
			issueRepo: issueRepo,
		}
		require.ErrorIs(t, r.RemoveRelationByID(ctx, relationID), ErrNotFound)
	})
}

func TestCachedIssueRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, issue *Issue) IssueRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts UpdateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Issue
		wantErr error
	}{
		{
			name: "update issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", issue.Parent.ID.String(), "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					allForProjectKeyCmd := new(redis.StringSliceCmd)
					allForProjectKeyCmd.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, forIssueKey).Return(forIssueKeyCmd)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, forIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				NumericID: 1,
				Parent: &PartialIssue{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					Assignments: make([]PartialAssignee, 0),
					Labels:      make([]PartialLabel, 0),
				},
				Kind:            model.IssueKindStory,
				Title:           "test issue",
				Description:     "test description",
				Status:          model.IssueStatusOpen,
				Priority:        model.IssuePriorityLow,
				Resolution:      model.IssueResolutionNone,
				ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Assignments:     make([]PartialAssignee, 0),
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
				WatcherCount:    convert.ToPointer(int64(0)),
				RelationCount:   convert.ToPointer(int64(0)),
				Links:           make([]model.IssueLink, 0),
			},
		},
		{
			name: "update issue with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *Issue) *redisBaseRepository {
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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, _ *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, IssueDetailProjection()).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				NumericID: 1,
				Parent: &PartialIssue{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					Assignments: make([]PartialAssignee, 0),
					Labels:      make([]PartialLabel, 0),
				},
				Kind:            model.IssueKindStory,
				Title:           "test issue",
				Description:     "test description",
				Status:          model.IssueStatusOpen,
				Priority:        model.IssuePriorityLow,
				Resolution:      model.IssueResolutionNone,
				ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Assignments:     make([]PartialAssignee, 0),
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
				WatcherCount:    convert.ToPointer(int64(0)),
				RelationCount:   convert.ToPointer(int64(0)),
				Links:           make([]model.IssueLink, 0),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update issue set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))

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
						Value: issue,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				NumericID: 1,
				Parent: &PartialIssue{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					Assignments: make([]PartialAssignee, 0),
					Labels:      make([]PartialLabel, 0),
				},
				Kind:            model.IssueKindStory,
				Title:           "test issue",
				Description:     "test description",
				Status:          model.IssueStatusOpen,
				Priority:        model.IssuePriorityLow,
				Resolution:      model.IssueResolutionNone,
				ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Assignments:     make([]PartialAssignee, 0),
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
				WatcherCount:    convert.ToPointer(int64(0)),
				RelationCount:   convert.ToPointer(int64(0)),
				Links:           make([]model.IssueLink, 0),
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update issue delete for issue to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", issue.Parent.ID.String(), "*")

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, forIssueKey).Return(forIssueKeyCmd)

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
					cacheRepo.EXPECT().Delete(ctx, forIssueKey).Return(assert.AnError)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				NumericID: 1,
				Parent: &PartialIssue{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					Assignments: make([]PartialAssignee, 0),
					Labels:      make([]PartialLabel, 0),
				},
				Kind:            model.IssueKindStory,
				Title:           "test issue",
				Description:     "test description",
				Status:          model.IssueStatusOpen,
				Priority:        model.IssuePriorityLow,
				Resolution:      model.IssueResolutionNone,
				ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Assignments:     make([]PartialAssignee, 0),
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
				WatcherCount:    convert.ToPointer(int64(0)),
				RelationCount:   convert.ToPointer(int64(0)),
				Links:           make([]model.IssueLink, 0),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update issue with delete projects cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(IssueDetailProjection()))
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", issue.Parent.ID.String(), "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					allForProjectKeyCmd := new(redis.StringSliceCmd)
					allForProjectKeyCmd.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, forIssueKey).Return(forIssueKeyCmd)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)
					cacheRepo.EXPECT().Delete(ctx, forIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts UpdateIssueOpts, issue *Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().Update(ctx, id, opts, IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:        model.MustNewID(model.ResourceTypeIssue),
				NumericID: 1,
				Parent: &PartialIssue{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					Assignments: make([]PartialAssignee, 0),
					Labels:      make([]PartialLabel, 0),
				},
				Kind:            model.IssueKindStory,
				Title:           "test issue",
				Description:     "test description",
				Status:          model.IssueStatusOpen,
				Priority:        model.IssuePriorityLow,
				Resolution:      model.IssueResolutionNone,
				ReportedBy:      &PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
				Assignments:     make([]PartialAssignee, 0),
				Labels:          make([]PartialLabel, 0),
				CommentCount:    convert.ToPointer(int64(0)),
				AttachmentCount: convert.ToPointer(int64(0)),
				WatcherCount:    convert.ToPointer(int64(0)),
				RelationCount:   convert.ToPointer(int64(0)),
				Links:           make([]model.IssueLink, 0),
			},
			wantErr: ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts, IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedIssueRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID) IssueRepository
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
	}{{
		name: "delete issue",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
				listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
				allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
				allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				listRelationsKeyResult := new(redis.StringSliceCmd)
				listRelationsKeyResult.SetVal([]string{listRelationsKey})

				allForIssueKeyResult := new(redis.StringSliceCmd)
				allForIssueKeyResult.SetVal([]string{allForIssueKey})

				allForProjectKeyResult := new(redis.StringSliceCmd)
				allForProjectKeyResult.SetVal([]string{allForProjectKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
				dbClient.EXPECT().Keys(ctx, listRelationsKey).Return(listRelationsKeyResult)
				dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
				dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(8)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(8)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				repo.EXPECT().Delete(ctx, id).Return(nil)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
	}, {
		name: "delete issue with deletion error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
				listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
				allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
				allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				listRelationsKeyResult := new(redis.StringSliceCmd)
				listRelationsKeyResult.SetVal([]string{listRelationsKey})

				allForIssueKeyResult := new(redis.StringSliceCmd)
				allForIssueKeyResult.SetVal([]string{allForIssueKey})

				allForProjectKeyResult := new(redis.StringSliceCmd)
				allForProjectKeyResult.SetVal([]string{allForProjectKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
				dbClient.EXPECT().Keys(ctx, listRelationsKey).Return(listRelationsKeyResult)
				dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
				dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(8)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(8)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				repo.EXPECT().Delete(ctx, id).Return(ErrNotFound)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrNotFound,
	}, {
		name: "delete issue with clear cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				return redisCacheExpectingPatterns(ctrl, ctx, []string{key}, 0, ErrCacheDelete)
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "delete issue with clear watchers cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

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
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "delete issue with clear relations cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(3)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "delete issue with clear for issue cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
				listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
				allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				listRelationsKeyResult := new(redis.StringSliceCmd)
				listRelationsKeyResult.SetVal([]string{listRelationsKey})

				allForIssueKeyResult := new(redis.StringSliceCmd)
				allForIssueKeyResult.SetVal([]string{allForIssueKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
				dbClient.EXPECT().Keys(ctx, listRelationsKey).Return(listRelationsKeyResult)
				dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(5)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "delete issue with clear for project cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
				listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
				allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
				allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				listRelationsKeyResult := new(redis.StringSliceCmd)
				listRelationsKeyResult.SetVal([]string{listRelationsKey})

				allForIssueKeyResult := new(redis.StringSliceCmd)
				allForIssueKeyResult.SetVal([]string{allForIssueKey})

				allForProjectKeyResult := new(redis.StringSliceCmd)
				allForProjectKeyResult.SetVal([]string{allForProjectKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
				dbClient.EXPECT().Keys(ctx, listRelationsKey).Return(listRelationsKeyResult)
				dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
				dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(6)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	}, {
		name: "delete issue with clear projects cache error",
		fields: fields{
			cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
				key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
				watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
				relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
				listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
				allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListForIssue", "*")
				allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*")
				getByKeyKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
				projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

				watchersKeyResult := new(redis.StringSliceCmd)
				watchersKeyResult.SetVal([]string{watchersKey})

				relationsKeyResult := new(redis.StringSliceCmd)
				relationsKeyResult.SetVal([]string{relationsKey})

				listRelationsKeyResult := new(redis.StringSliceCmd)
				listRelationsKeyResult.SetVal([]string{listRelationsKey})

				allForIssueKeyResult := new(redis.StringSliceCmd)
				allForIssueKeyResult.SetVal([]string{allForIssueKey})

				allForProjectKeyResult := new(redis.StringSliceCmd)
				allForProjectKeyResult.SetVal([]string{allForProjectKey})

				getByKeyKeyResult := new(redis.StringSliceCmd)
				getByKeyKeyResult.SetVal([]string{getByKeyKey})

				projectsKeyResult := new(redis.StringSliceCmd)
				projectsKeyResult.SetVal([]string{projectsKey})

				keyCmd := new(redis.StringSliceCmd)
				keyCmd.SetVal([]string{key})

				dbClient := mock.NewUniversalClient(ctrl)
				dbClient.EXPECT().Keys(ctx, key).Return(keyCmd)
				dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
				dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
				dbClient.EXPECT().Keys(ctx, listRelationsKey).Return(listRelationsKeyResult)
				dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
				dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
				dbClient.EXPECT().Keys(ctx, getByKeyKey).Return(getByKeyKeyResult)
				dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

				db, err := NewRedisDatabase(
					WithRedisClient(dbClient),
				)
				require.NoError(t, err)

				span := mock.NewMockSpan(ctrl)
				span.EXPECT().End(gomock.Len(0)).Times(8)

				tracer := mock.NewMockTracer(ctrl)
				tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(8)

				cacheRepo := mock.NewCacheBackend(ctrl)
				cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, getByKeyKey).Return(nil)
				cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)

				return &redisBaseRepository{
					db:     db,
					cache:  cacheRepo,
					tracer: tracer,
					logger: mock.NewMockLogger(ctrl),
				}
			},
			issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID) IssueRepository {
				repo := NewMockIssueRepository(ctrl)
				return repo
			},
		},
		args: args{
			ctx: context.Background(),
			id:  model.MustNewID(model.ResourceTypeIssue),
		},
		wantErr: ErrCacheDelete,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDecodeIssueLinks(t *testing.T) {
	t.Parallel()

	t.Run("splits encoded url and label", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links": []any{
				"https://example.com/a" + issueLinkLabelSep + "Spec",
				"https://example.com/b" + issueLinkLabelSep + "Runbook",
			},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/a", Label: "Spec"},
			{URL: "https://example.com/b", Label: "Runbook"},
		}, got)
	})

	t.Run("defaults missing labels to the url", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links": []string{"https://example.com/legacy"},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/legacy", Label: "https://example.com/legacy"},
		}, got)
	})

	t.Run("zips leftover parallel label lists", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links":       []any{"https://example.com/a", "https://example.com/b"},
			"link_labels": []any{"Spec", "Runbook"},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/a", Label: "Spec"},
			{URL: "https://example.com/b", Label: "Runbook"},
		}, got)
	})

	t.Run("returns an empty slice when links are absent", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{})
		assert.Equal(t, []model.IssueLink{}, got)
	})
}

func TestEncodeIssueLinks(t *testing.T) {
	t.Parallel()

	got := encodeIssueLinks([]model.IssueLink{
		{URL: "https://example.com/a", Label: "Spec"},
		{URL: "https://example.com/b", Label: "https://example.com/b"},
		{URL: "https://example.com/c", Label: ""},
	})
	assert.Equal(t, []string{
		"https://example.com/a" + issueLinkLabelSep + "Spec",
		"https://example.com/b",
		"https://example.com/c",
	}, got)

	assert.Equal(t, []string{}, encodeIssueLinks(nil))
}

func TestUpdateIssueOpts_patch(t *testing.T) {
	t.Parallel()

	t.Run("nil description clears the field", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{
			Description: optional.Null[string](),
		}.patch()
		require.Contains(t, got, "description")
		assert.Nil(t, got["description"])
	})

	t.Run("set description", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{
			Description: optional.Some("updated description"),
		}.patch()
		assert.Equal(t, "updated description", got["description"])
	})

	t.Run("undefined description is omitted", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{}.patch()
		_, ok := got["description"]
		assert.False(t, ok)
	})
}

func TestParentProjectKeyFromRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		record   *neo4j.Record
		fallback string
		want     string
	}{
		{
			name:     "missing field uses fallback",
			record:   &neo4j.Record{Keys: []string{}, Values: []any{}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "nil value uses fallback",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{nil}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "empty string uses fallback",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{""}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "parent project key wins",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{"PLAT"}},
			fallback: "ENG",
			want:     "PLAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, parentProjectKeyFromRecord(tt.record, tt.fallback))
		})
	}
}

func TestRelationCountAfterCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		hasParent bool
		want      int64
	}{
		{
			name:      "without parent",
			hasParent: false,
			want:      0,
		},
		{
			name:      "with parent",
			hasParent: true,
			want:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := relationCountAfterCreate(tt.hasParent)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestNeo4jIssueRepository_applyIssueLoadersUnknown(t *testing.T) {
	t.Parallel()

	r := new(Neo4jIssueRepository)
	err := r.applyIssueLoaders(context.Background(), nil, QueryPlan{
		Loaders: []CompiledQuery{{Name: "issue.load_unknown", Params: map[string]any{}}},
	}, []*issueDetailRow{{
		projectKey: "ENG",
		issue:      &Issue{ID: model.MustNewID(model.ResourceTypeIssue)},
	}})
	assert.ErrorIs(t, err, ErrQueryCompile)
}
