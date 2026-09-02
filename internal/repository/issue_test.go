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

func issueRelationPairCachePatterns(source, target model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeIssue.String(), "Get", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", source.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "Get", target.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", target.String(), "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", target.String(), "*"),
	}
}

func TestCachedIssueRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) repository.IssueRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "creates issue and bumps scoped generations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) []repository.RedisRepositoryOption {
					projectsPattern := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectGenKey := issueListProjectGenKey(opts.ProjectID)

					projectsCmd := new(redis.StringSliceCmd)
					projectsCmd.SetVal([]string{projectsPattern})
					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, projectsPattern).Return(projectsCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, projectGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsPattern).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Issue{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueOpts{
					ProjectID:   model.MustNewID(model.ResourceTypeProject),
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
		},
		{
			name: "returns create error without cache writes",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateIssueOpts) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)
					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(nil, repository.ErrIssueCreate)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueOpts{
					ProjectID:   model.MustNewID(model.ResourceTypeProject),
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrIssueCreate,
		},
		{
			name: "returns generation bump error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) []repository.RedisRepositoryOption {
					projectGenKey := issueListProjectGenKey(opts.ProjectID)
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)
					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, projectGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectGenKey, Value: int64(1)}).Return(repository.ErrCacheWrite)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Issue{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueOpts{
					ProjectID:   model.MustNewID(model.ResourceTypeProject),
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "returns cross-cache clear error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) []repository.RedisRepositoryOption {
					projectGenKey := issueListProjectGenKey(opts.ProjectID)
					projectsPattern := composeCacheKey(model.ResourceTypeProject.String(), "*")

					projectsCmd := new(redis.StringSliceCmd)
					projectsCmd.SetVal([]string{projectsPattern})
					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, projectsPattern).Return(projectsCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)
					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, projectGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectsPattern).Return(repository.ErrCacheDelete)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Create(ctx, opts).Return(&repository.Issue{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueOpts{
					ProjectID:   model.MustNewID(model.ResourceTypeProject),
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.opts),
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

func TestCachedIssueRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *repository.Issue
		wantErr error
	}{
		{
			name: "get uncached issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))

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
						Value: issue,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *repository.Issue {
				return &repository.Issue{
					ID:              id,
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))

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
						if ptr, ok := dst.(**repository.Issue); ok {
							*ptr = issue
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Issue) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *repository.Issue {
				return &repository.Issue{
					ID:              id,
					NumericID:       1,
					Parent:          nil,
					Kind:            model.IssueKindStory,
					Title:           "test issue",
					Description:     "test description",
					Status:          model.IssueStatusOpen,
					Priority:        model.IssuePriorityLow,
					Resolution:      model.IssueResolutionNone,
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached issue error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))

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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Issue) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached issue cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))

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
						Value: issue,
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			var want *repository.Issue
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, repository.IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedIssueRepository_GetByKey(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *repository.Issue) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *repository.Issue) repository.IssueRepository
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
		want    *repository.Issue
		wantErr error
	}{
		{
			name: "get uncached issue by key",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", namespaceID.String(), issueKey, projectionCacheValue(repository.IssueDetailProjection()))

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
						Value: issue,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetByKey(ctx, namespaceID, issueKey, repository.IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				issueKey:    "ENG-42",
			},
			want: &repository.Issue{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, issueKey string, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", namespaceID.String(), issueKey, projectionCacheValue(repository.IssueDetailProjection()))

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
						if ptr, ok := dst.(**repository.Issue); ok {
							*ptr = issue
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ string, _ *repository.Issue) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				issueKey:    "ENG-42",
			},
			want: &repository.Issue{
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

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.GetByKey(tt.args.ctx, tt.args.namespaceID, tt.args.issueKey, repository.IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForProject(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository
	}
	type args struct {
		ctx   context.Context
		query repository.IssueListQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    repository.Page[*repository.PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListQuery, key string, _ repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectGenKey(query.ProjectID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForProject(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListQuery{
					ProjectID:  model.MustNewID(model.ResourceTypeProject),
					Page:       repository.CursorPage{Size: 10},
					Projection: repository.IssueListForProjectProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "PROJ-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectGenKey(query.ProjectID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						*(dst.(*repository.Page[*repository.PartialIssue])) = page
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.IssueListQuery, _ repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListQuery{
					ProjectID:  model.MustNewID(model.ResourceTypeProject),
					Page:       repository.CursorPage{Size: 10},
					Projection: repository.IssueListForProjectProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "PROJ-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			baseKey := mustPlanCacheKey(t, tt.args.query, model.ResourceTypeIssue.String(), "ListForProject", tt.args.query.ProjectID.String())
			key := baseKey + ":g:0:ae:0:pe:0"

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.query, key, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForProject(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForNamespace(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForNamespaceQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForNamespaceQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository
	}
	type args struct {
		ctx   context.Context
		query repository.IssueListForNamespaceQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    repository.Page[*repository.PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForNamespaceQuery, key string, _ repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListNamespaceGenKey(query.NamespaceID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForNamespaceQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForNamespace(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListForNamespaceQuery{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					Page:        repository.CursorPage{Size: 10},
					Projection:  repository.IssueListForNamespaceProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForNamespaceQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListNamespaceGenKey(query.NamespaceID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						*(dst.(*repository.Page[*repository.PartialIssue])) = page
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.IssueListForNamespaceQuery, _ repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListForNamespaceQuery{
					NamespaceID: model.MustNewID(model.ResourceTypeNamespace),
					Page:        repository.CursorPage{Size: 10},
					Projection:  repository.IssueListForNamespaceProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			baseKey := mustPlanCacheKey(t, tt.args.query, model.ResourceTypeIssue.String(), "ListForNamespace", tt.args.query.NamespaceID.String())
			key := baseKey + ":g:0:ae:0:pe:0"

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.query, key, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForNamespace(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForUser(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForUserQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForUserQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository
	}
	type args struct {
		ctx   context.Context
		query repository.IssueListForUserQuery
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    repository.Page[*repository.PartialIssue]
		wantErr error
	}{
		{
			name: "get uncached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForUserQuery, key string, _ repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListUserGenKey(query.UserID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(gomock.Any()).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForUserQuery, page repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForUser(ctx, query).Return(page, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListForUserQuery{
					UserID:     model.MustNewID(model.ResourceTypeUser),
					Page:       repository.CursorPage{Size: 10},
					Projection: repository.IssueListForUserProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "test issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
		{
			name: "get cached issue page",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, query repository.IssueListForUserQuery, key string, page repository.Page[*repository.PartialIssue]) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, issueListUserGenKey(query.UserID), gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						*(dst.(*repository.Page[*repository.PartialIssue])) = page
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.IssueListForUserQuery, _ repository.Page[*repository.PartialIssue]) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				query: repository.IssueListForUserQuery{
					UserID:     model.MustNewID(model.ResourceTypeUser),
					Page:       repository.CursorPage{Size: 10},
					Projection: repository.IssueListForUserProjection(),
				},
			},
			want: repository.Page[*repository.PartialIssue]{
				Items: []*repository.PartialIssue{
					{
						ID:          model.MustNewID(model.ResourceTypeIssue),
						Key:         "ENG-1",
						NumericID:   1,
						Kind:        model.IssueKindStory,
						Title:       "cached issue",
						Status:      model.IssueStatusOpen,
						Priority:    model.IssuePriorityLow,
						Assignments: make([]repository.PartialAssignee, 0),
						Labels:      make([]repository.PartialLabel, 0),
					},
				},
				PageInfo: repository.PageInfo{HasMore: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			baseKey := mustPlanCacheKey(t, tt.args.query, model.ResourceTypeIssue.String(), "ListForUser", tt.args.query.UserID.String())
			key := baseKey + ":g:0:ae:0:pe:0"

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.query, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.query, key, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForUser(tt.args.ctx, tt.args.query)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_ListForIssue(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) repository.IssueRepository
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
		want    []*repository.Issue
		wantErr error
	}{
		{
			name: "get uncached issues",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) []repository.RedisRepositoryOption {
					key := issueListForIssueCacheKey(issue, repository.CursorPage{Size: testPageSize(limit)}, repository.IssueDetailProjection())

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
						Value: repository.Page[*repository.Issue]{Items: issues},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, repository.IssueListForIssueQuery{IssueID: issue, Page: repository.CursorPage{Size: limit}, Projection: repository.IssueDetailProjection()}).Return(repository.Page[*repository.Issue]{Items: issues}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.Issue{
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
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) []repository.RedisRepositoryOption {
					key := issueListForIssueCacheKey(issue, repository.CursorPage{Size: testPageSize(limit)}, repository.IssueDetailProjection())

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
						if ptr, ok := dst.(*repository.Page[*repository.Issue]); ok {
							*ptr = repository.Page[*repository.Issue]{Items: issues}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Issue) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.Issue{
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
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
					ReportedBy:      &repository.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
					Assignments:     make([]repository.PartialAssignee, 0),
					Labels:          make([]repository.PartialLabel, 0),
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*repository.Issue) []repository.RedisRepositoryOption {
					key := issueListForIssueCacheKey(issue, repository.CursorPage{Size: testPageSize(limit)}, repository.IssueDetailProjection())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, repository.IssueListForIssueQuery{IssueID: issue, Page: repository.CursorPage{Size: limit}, Projection: repository.IssueDetailProjection()}).Return(repository.Page[*repository.Issue]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get issues cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, _ []*repository.Issue) []repository.RedisRepositoryOption {
					key := issueListForIssueCacheKey(issue, repository.CursorPage{Size: testPageSize(limit)}, repository.IssueDetailProjection())

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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _, _ int, _ []*repository.Issue) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached issues cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) []repository.RedisRepositoryOption {
					key := issueListForIssueCacheKey(issue, repository.CursorPage{Size: testPageSize(limit)}, repository.IssueDetailProjection())

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
						Value: repository.Page[*repository.Issue]{Items: issues},
					}).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, _, limit int, issues []*repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().ListForIssue(ctx, repository.IssueListForIssueQuery{IssueID: issue, Page: repository.CursorPage{Size: limit}, Projection: repository.IssueDetailProjection()}).Return(repository.Page[*repository.Issue]{Items: issues}, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, testPageSize(tt.args.limit), tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, testPageSize(tt.args.limit), tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListForIssue(tt.args.ctx, repository.IssueListForIssueQuery{IssueID: tt.args.issue, Page: repository.CursorPage{Size: testPageSize(tt.args.limit)}, Projection: repository.IssueDetailProjection()})
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedIssueRepository_AddWatcher(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().AddWatcher(ctx, id, watcher).Return(nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeIssue),
				watcher: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "add watcher with repository error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().AddWatcher(ctx, id, watcher).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeIssue),
				watcher: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "add watcher with clear issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, 0, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add watcher with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, 1, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.AddWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetWatchers(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.User
		wantErr error
	}{
		{
			name: "get issue watchers",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						Value: watchers,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(watchers, nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.User{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.User) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get issue watchers from cache",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						if ptr, ok := dst.(*[]*repository.User); ok {
							*ptr = watchers
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*repository.User) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.User{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
						Value: watchers,
					}).Return(repository.ErrCacheWrite)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, watchers []*repository.User) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetWatchers(ctx, id).Return(watchers, nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.User{
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "get issue watchers with get cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.User) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(repository.ErrCacheRead)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*repository.User) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.User{
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
			wantErr: repository.ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveWatcher(ctx, id, watcher).Return(nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeIssue),
				watcher: model.MustNewID(model.ResourceTypeUser),
			},
		},
		{
			name: "remove issue watcher with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveWatcher(ctx, id, watcher).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				id:      model.MustNewID(model.ResourceTypeIssue),
				watcher: model.MustNewID(model.ResourceTypeUser),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "remove issue watcher with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, 0, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id, _ model.ID) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
					}, 1, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.watcher)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.RemoveWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_AddRelation(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) repository.IssueRepository
	}
	type args struct {
		ctx  context.Context
		opts repository.CreateIssueRelationOpts
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().AddRelation(ctx, opts).Return(&repository.IssueRelation{}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "add issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().AddRelation(ctx, opts).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "add issue relation with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts repository.CreateIssueRelationOpts) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(opts.Source, opts.Target), 0, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ repository.CreateIssueRelationOpts) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				opts: repository.CreateIssueRelationOpts{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.opts),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.opts)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			_, err := r.AddRelation(tt.args.ctx, tt.args.opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetRelation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)
	want := &repository.IssueRelation{ID: relationID, Kind: model.IssueRelationKindBlocks}

	ctrl := gomock.NewController(t)

	repo := mockrepo.NewMockIssueRepository(ctrl)
	repo.EXPECT().GetRelation(ctx, relationID).Return(want, nil)

	r := func() *repository.RedisCachedIssueRepository {
		r, err := repository.NewCachedIssueRepository(
			repo,
			redisCacheExpectingPatterns(ctrl, ctx, nil, -1, nil)...,
		)
		if err != nil {
			panic(err)
		}
		return r
	}()

	got, err := r.GetRelation(ctx, relationID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCachedIssueRepository_ListRelations(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	query := repository.IssueRelationListQuery{IssueID: issueID, Page: repository.CursorPage{Size: 10}}
	key := mustPlanCacheKey(t, query, model.ResourceTypeIssue.String(), "ListRelations", issueID.String())
	page := repository.Page[*repository.IssueRelationItem]{Items: []*repository.IssueRelationItem{{ID: model.MustNewID(model.ResourceTypeIssueRelation)}}}

	t.Run("get uncached relation page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(2)
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Get", gomock.Len(0)).Return(context.Background(), span)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Set", gomock.Len(0)).Return(context.Background(), span)

		cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(gomock.Any(), key, gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Set(&cache.Item{
			Ctx:   context.Background(),
			Key:   key,
			Value: page,
		}).Return(nil)

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().ListRelations(gomock.Any(), query).Return(page, nil)

		r := func() *repository.RedisCachedIssueRepository {
			r, err := repository.NewCachedIssueRepository(
				issueRepo,
				[]repository.RedisRepositoryOption{
					repository.WithRedisDatabase(db),
					repository.WithCacheBackend(cacheRepo),
					repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
					repository.WithRedisRepositoryTracer(tracer),
				}...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListRelations(context.Background(), query)
		require.NoError(t, err)
		assert.Equal(t, page, got)
	})

	t.Run("get cached relation page", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "repository.redisBaseRepository/Get", gomock.Len(0)).Return(context.Background(), span)

		cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(gomock.Any(), key, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest any) error {
			*(dest.(*repository.Page[*repository.IssueRelationItem])) = page
			return nil
		})

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)

		r := func() *repository.RedisCachedIssueRepository {
			r, err := repository.NewCachedIssueRepository(
				mockrepo.NewMockIssueRepository(ctrl),
				[]repository.RedisRepositoryOption{
					repository.WithRedisDatabase(db),
					repository.WithCacheBackend(cacheRepo),
					repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
					repository.WithRedisRepositoryTracer(tracer),
				}...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListRelations(context.Background(), query)
		require.NoError(t, err)
		assert.Equal(t, page, got)
	})
}

func TestCachedIssueRepository_GetRelations(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*repository.IssueRelation
		wantErr error
	}{
		{
			name: "get issue relations",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						Value: relations,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(relations, nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.IssueRelation{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.IssueRelation) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.IssueRelation) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get issue relations from cache",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						if ptr, ok := dst.(*[]*repository.IssueRelation); ok {
							*ptr = relations
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*repository.IssueRelation) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.IssueRelation{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

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
						Value: relations,
					}).Return(repository.ErrCacheWrite)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, relations []*repository.IssueRelation) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().GetRelations(ctx, id).Return(relations, nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.IssueRelation{
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "get issue relations with get cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ []*repository.IssueRelation) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(repository.ErrCacheRead)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ []*repository.IssueRelation) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*repository.IssueRelation{
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
			wantErr: repository.ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveRelation(ctx, source, target, kind).Return(nil).Times(1)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveRelation(ctx, source, target, kind).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeIssue),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "remove issue relation with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, _ model.IssueRelationKind) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), 0, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) repository.IssueRepository {
					return mockrepo.NewMockIssueRepository(ctrl)
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeIssue),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
	rel := &repository.IssueRelation{ID: relationID, Source: source, Target: target, Kind: model.IssueRelationKindBlocks}

	t.Run("remove relation by id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(ctx, relationID).Return(rel, nil)
		issueRepo.EXPECT().RemoveRelationByID(ctx, relationID).Return(nil)

		r := func() *repository.RedisCachedIssueRepository {
			r, err := repository.NewCachedIssueRepository(
				issueRepo,
				redisCacheExpectingPatterns(ctrl, ctx, issueRelationPairCachePatterns(source, target), -1, nil)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.NoError(t, r.RemoveRelationByID(ctx, relationID))
	})

	t.Run("get relation error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(ctx, relationID).Return(nil, repository.ErrNotFound)

		r := func() *repository.RedisCachedIssueRepository {
			r, err := repository.NewCachedIssueRepository(
				issueRepo,
				redisCacheExpectingPatterns(ctrl, ctx, nil, -1, nil)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		require.ErrorIs(t, r.RemoveRelationByID(ctx, relationID), repository.ErrNotFound)
	})
}

func TestCachedIssueRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateIssueOpts, issue *repository.Issue) repository.IssueRepository
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts repository.UpdateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *repository.Issue
		wantErr error
	}{
		{
			name: "update issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))
					projectPattern := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectGenKey := issueListProjectGenKey(issue.Project.ID)
					assigneeGenKey := issueListUserGenKey(issue.Assignments[0].ID)

					projectPatternCmd := new(redis.StringSliceCmd)
					projectPatternCmd.SetVal([]string{projectPattern})
					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, projectPattern).Return(projectPatternCmd)

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(3)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: issue}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, assigneeGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: assigneeGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectPattern).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateIssueOpts, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(&repository.Issue{}, nil)
					repo.EXPECT().Update(ctx, id, opts, repository.IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				opts: repository.UpdateIssueOpts{
					Title: optional.Some("updated title"),
				},
			},
			want: &repository.Issue{
				ID:      model.MustNewID(model.ResourceTypeIssue),
				Project: &repository.PartialProject{ID: model.MustNewID(model.ResourceTypeProject)},
				Assignments: []repository.PartialAssignee{
					{ID: model.MustNewID(model.ResourceTypeUser), Kind: model.AssignmentKindAssignee},
				},
			},
		},
		{
			name: "update issue with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ *repository.Issue) []repository.RedisRepositoryOption {
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)
					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateIssueOpts, _ *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(nil, nil)
					repo.EXPECT().Update(ctx, id, opts, repository.IssueDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update issue set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(repository.IssueDetailProjection()))
					db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)
					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: key, Value: issue}).Return(repository.ErrCacheWrite)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts repository.UpdateIssueOpts, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(nil, nil)
					repo.EXPECT().Update(ctx, id, opts, repository.IssueDetailProjection()).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: &repository.Issue{
				ID:      model.MustNewID(model.ResourceTypeIssue),
				Project: &repository.PartialProject{ID: model.MustNewID(model.ResourceTypeProject)},
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.opts, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts, repository.IssueDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedIssueRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository
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
			name: "delete issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) []repository.RedisRepositoryOption {
					getKey := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*")
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					listRelationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*")
					getByKeyPattern := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*")
					projectPattern := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectGenKey := issueListProjectGenKey(issue.Project.ID)
					assigneeGenKey := issueListUserGenKey(issue.Assignments[0].ID)

					dbClient := mockrepo.NewMockUniversalClient(ctrl)
					for _, pattern := range []string{getKey, watchersKey, relationsKey, listRelationsKey, getByKeyPattern, projectPattern} {
						cmd := new(redis.StringSliceCmd)
						cmd.SetVal([]string{pattern})
						dbClient.EXPECT().Keys(ctx, pattern).Return(cmd)
					}

					db, err := repository.NewRedisDatabase(repository.WithRedisClient(dbClient))
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(10)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(6)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, getKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, listRelationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, getByKeyPattern).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, projectPattern).Return(nil)
					cacheRepo.EXPECT().Get(ctx, projectGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: projectGenKey, Value: int64(1)}).Return(nil)
					cacheRepo.EXPECT().Get(ctx, assigneeGenKey, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{Ctx: ctx, Key: assigneeGenKey, Value: int64(1)}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(issue, nil)
					repo.EXPECT().Delete(ctx, id).Return(nil).Times(1)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: nil,
		},
		{
			name: "delete issue with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Issue) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListRelations", id.String(), "*"),
						composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", "*"),
						composeCacheKey(model.ResourceTypeProject.String(), "*"),
					}, -1, nil)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(issue, nil)
					repo.EXPECT().Delete(ctx, id).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "delete issue with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Issue) []repository.RedisRepositoryOption {
					return redisCacheExpectingPatterns(ctrl, ctx, []string{
						composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), "*"),
					}, 0, repository.ErrCacheDelete)
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *repository.Issue) repository.IssueRepository {
					repo := mockrepo.NewMockIssueRepository(ctrl)
					repo.EXPECT().Get(ctx, id, repository.IssueProjection{Assignments: true}).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			issue := &repository.Issue{
				Project: &repository.PartialProject{ID: model.MustNewID(model.ResourceTypeProject)},
				Assignments: []repository.PartialAssignee{
					{ID: model.MustNewID(model.ResourceTypeUser), Kind: model.AssignmentKindAssignee},
				},
			}

			r := func() *repository.RedisCachedIssueRepository {
				r, err := repository.NewCachedIssueRepository(
					tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.id, issue),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, issue)...,
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
