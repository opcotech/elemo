package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
	}{
		{
			name: "add new issue with no parent",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)

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
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "add new issue with parent",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")
					parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", opts.Parent.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					parentIssueKeyResult := new(redis.StringSliceCmd)
					parentIssueKeyResult.SetVal([]string{parentIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
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
					cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)
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
					NumericID:   1,
					Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "add new issue with error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allProjectsKey).Return(nil)

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
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
				},
			},
			wantErr: ErrIssueCreate,
		},
		{
			name: "add new issue with cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, allProjectsKey).Return(allProjectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
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
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new issue with parent issue cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")
					parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", opts.Parent.String(), "*")

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					parentIssueKeyResult := new(redis.StringSliceCmd)
					parentIssueKeyResult.SetVal([]string{parentIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)
					dbClient.EXPECT().Keys(ctx, parentIssueKey).Return(parentIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
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
					NumericID:   1,
					Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
				},
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "add new issue with project cache delete error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueOpts) *redisBaseRepository {
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", opts.ProjectID.String(), "*")

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
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Links:       make([]string, 0),
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					repo.EXPECT().Get(ctx, id).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *Issue {
				return &Issue{
					ID:          id,
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				}
			},
		},
		{
			name: "get cached issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					ID:          id,
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				}
			},
		},
		{
			name: "get uncached issue error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					repo.EXPECT().Get(ctx, id).Return(nil, ErrNotFound)
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					repo.EXPECT().Get(ctx, id).Return(issue, nil)
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
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedIssueRepository_GetAllForProject(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) IssueRepository
	}
	type args struct {
		ctx     context.Context
		project model.ID
		offset  int
		limit   int
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
						Value: issues,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForProject(ctx, project, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			want: []*Issue{
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "get cached issues",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
						if ptr, ok := dst.(*[]*Issue); ok {
							*ptr = issues
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
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			want: []*Issue{
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "get uncached issues error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, _ []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, _ []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForProject(ctx, project, offset, limit).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get get issues cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, _ []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: ErrCacheRead,
		},
		{
			name: "get uncached issues cache set error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
						Value: issues,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, project model.ID, offset, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForProject(ctx, project, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
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
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllForProject(tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_GetAllForIssue(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) *redisBaseRepository
		issueRepo func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) IssueRepository
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
						Value: issues,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForIssue(ctx, issue, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*Issue{
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "get cached issues",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
						if ptr, ok := dst.(*[]*Issue); ok {
							*ptr = issues
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
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
				{
					ID:          model.MustNewID(model.ResourceTypeIssue),
					NumericID:   1,
					Parent:      nil,
					Kind:        model.IssueKindStory,
					Title:       "test issue",
					Description: "test description",
					Status:      model.IssueStatusOpen,
					Priority:    model.IssuePriorityLow,
					Resolution:  model.IssueResolutionNone,
					ReportedBy:  model.MustNewID(model.ResourceTypeUser),
					Assignees:   make([]model.ID, 0),
					Labels:      make([]model.ID, 0),
					Comments:    make([]model.ID, 0),
					Attachments: make([]model.ID, 0),
					Watchers:    make([]model.ID, 0),
					Relations:   make([]model.ID, 0),
					Links:       make([]string, 0),
				},
			},
		},
		{
			name: "get uncached issues error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, _ []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, _ []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForIssue(ctx, issue, offset, limit).Return(nil, ErrNotFound)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, _ []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
						Value: issues,
					}).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, issue model.ID, offset, limit int, issues []*Issue) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().GetAllForIssue(ctx, issue, offset, limit).Return(issues, nil)
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
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit, tt.want),
				issueRepo: tt.fields.issueRepo(ctrl, tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllForIssue(tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
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
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

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
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
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
			name: "add issue relation non-issue relation",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Target.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Target.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
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
					Source: model.MustNewID(model.ResourceTypeDocument),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "add issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
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
			wantErr: ErrCacheDelete,
		},
		{
			name: "add issue relation with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Source.String(), "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
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
			wantErr: ErrCacheDelete,
		},
		{
			name: "add issue relation with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
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
			wantErr: ErrCacheDelete,
		},
		{
			name: "add issue relation with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, opts CreateIssueRelationOpts) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), opts.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", opts.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _ CreateIssueRelationOpts) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
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
			name: "remove issue relation non-issue relation",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, _, target model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), target.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", target.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, ctx context.Context, source, target model.ID, kind model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					repo.EXPECT().RemoveRelation(ctx, source, target, kind).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:    context.Background(),
				source: model.MustNewID(model.ResourceTypeDocument),
				target: model.MustNewID(model.ResourceTypeIssue),
				kind:   model.IssueRelationKindBlocks,
			},
		},
		{
			name: "remove issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
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
		{
			name: "remove issue relation with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
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
		{
			name: "remove issue relation with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
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
		{
			name: "remove issue relation with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(ErrCacheDelete)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				issueRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ model.IssueRelationKind) IssueRepository {
					repo := NewMockIssueRepository(ctrl)
					return repo
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
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, forIssueKey).Return(forIssueKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, forIssueKey).Return(nil)
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
					repo.EXPECT().Update(ctx, id, opts).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:          model.MustNewID(model.ResourceTypeIssue),
				NumericID:   1,
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Assignees:   make([]model.ID, 0),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
				Watchers:    make([]model.ID, 0),
				Relations:   make([]model.ID, 0),
				Links:       make([]string, 0),
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
					repo.EXPECT().Update(ctx, id, opts).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:          model.MustNewID(model.ResourceTypeIssue),
				NumericID:   1,
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Assignees:   make([]model.ID, 0),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
				Watchers:    make([]model.ID, 0),
				Relations:   make([]model.ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "update issue set cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
					repo.EXPECT().Update(ctx, id, opts).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:          model.MustNewID(model.ResourceTypeIssue),
				NumericID:   1,
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Assignees:   make([]model.ID, 0),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
				Watchers:    make([]model.ID, 0),
				Relations:   make([]model.ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrCacheWrite,
		},
		{
			name: "update issue delete for issue to cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

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
					repo.EXPECT().Update(ctx, id, opts).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:          model.MustNewID(model.ResourceTypeIssue),
				NumericID:   1,
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Assignees:   make([]model.ID, 0),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
				Watchers:    make([]model.ID, 0),
				Relations:   make([]model.ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrCacheDelete,
		},
		{
			name: "update issue with delete projects cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, issue *Issue) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, forIssueKey).Return(forIssueKeyCmd)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyCmd)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, projectsKey).Return(ErrCacheDelete)
					cacheRepo.EXPECT().Delete(ctx, forIssueKey).Return(nil)
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
					repo.EXPECT().Update(ctx, id, opts).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.MustNewID(model.ResourceTypeIssue),
				opts: UpdateIssueOpts{},
			},
			want: &Issue{
				ID:          model.MustNewID(model.ResourceTypeIssue),
				NumericID:   1,
				Parent:      convert.ToPointer(model.MustNewID(model.ResourceTypeIssue)),
				Kind:        model.IssueKindStory,
				Title:       "test issue",
				Description: "test description",
				Status:      model.IssueStatusOpen,
				Priority:    model.IssuePriorityLow,
				Resolution:  model.IssueResolutionNone,
				ReportedBy:  model.MustNewID(model.ResourceTypeUser),
				Assignees:   make([]model.ID, 0),
				Labels:      make([]model.ID, 0),
				Comments:    make([]model.ID, 0),
				Attachments: make([]model.ID, 0),
				Watchers:    make([]model.ID, 0),
				Relations:   make([]model.ID, 0),
				Links:       make([]string, 0),
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
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.opts)
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
	}{
		{
			name: "delete issue",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
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
		},
		{
			name: "delete issue with deletion error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
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
		},
		{
			name: "delete issue with clear cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
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
		{
			name: "delete issue with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)

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
		},
		{
			name: "delete issue with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(3)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(2)

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
		},
		{
			name: "delete issue with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(4)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(3)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
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
		},
		{
			name: "delete issue with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(5)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(4)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
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
		},
		{
			name: "delete issue with clear projects cache error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := mock.NewUniversalClient(ctrl)
					dbClient.EXPECT().Keys(ctx, watchersKey).Return(watchersKeyResult)
					dbClient.EXPECT().Keys(ctx, relationsKey).Return(relationsKeyResult)
					dbClient.EXPECT().Keys(ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.EXPECT().Keys(ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.EXPECT().Keys(ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewRedisDatabase(
						WithRedisClient(dbClient),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(6)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Delete", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(5)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Delete(ctx, key).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, watchersKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, relationsKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForIssueKey).Return(nil)
					cacheRepo.EXPECT().Delete(ctx, allForProjectKey).Return(nil)
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
