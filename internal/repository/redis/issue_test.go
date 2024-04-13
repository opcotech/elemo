package redis

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedIssueRepository_Create(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository
		issueRepo func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository
	}
	type args struct {
		ctx     context.Context
		project model.ID
		issue   *model.Issue
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
				cacheRepo: func(ctx context.Context, project model.ID, _ *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Create", ctx, project, issue).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
			name: "add new issue with parent",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")
					parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					parentIssueKeyResult := new(redis.StringSliceCmd)
					parentIssueKeyResult.SetVal([]string{parentIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, parentIssueKey).Return(parentIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, parentIssueKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Create", ctx, project, issue).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
		},
		{
			name: "add new issue with error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, _ *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Create", ctx, project, issue).Return(repository.ErrIssueCreate)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
			wantErr: repository.ErrIssueCreate,
		},
		{
			name: "add new issue with cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, _ *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ *model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new issue with parent issue cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")
					parentIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					parentIssueKeyResult := new(redis.StringSliceCmd)
					parentIssueKeyResult.SetVal([]string{parentIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, parentIssueKey).Return(parentIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, parentIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Create", ctx, project, issue).Return(nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add new issue with project cache delete error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, _ *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ *model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
				issue: &model.Issue{
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.project, tt.args.issue),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.project, tt.args.issue),
			}
			err := r.Create(tt.args.ctx, tt.args.project, tt.args.issue)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository
		issueRepo func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *model.Issue
		wantErr error
	}{
		{
			name: "get uncached issue",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Get", ctx, id).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *model.Issue {
				return &model.Issue{
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
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issue, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ *model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *model.Issue {
				return &model.Issue{
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
				cacheRepo: func(ctx context.Context, id model.ID, _ *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, _ *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Get", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, _ *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ *model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Get", ctx, id).Return(issue, nil)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want *model.Issue
			if tt.want != nil {
				want = tt.want(tt.args.id)
			}

			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, want),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, want, got)
		})
	}
}

func TestCachedIssueRepository_GetAllForProject(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository
		issueRepo func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository
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
		want    []*model.Issue
		wantErr error
	}{
		{
			name: "get uncached issues",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issues,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForProject", ctx, project, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			want: []*model.Issue{
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
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issues, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _, _ int, _ []*model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			want: []*model.Issue{
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
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, _ []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, _ []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForProject", ctx, project, offset, limit).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get get issues cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, _ []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _, _ int, _ []*model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheRead,
		},
		{
			name: "get uncached issues cache set error",
			fields: fields{
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issues,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForProject", ctx, project, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:     context.Background(),
				project: model.MustNewID(model.ResourceTypeProject),
			},
			wantErr: repository.ErrCacheWrite,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit, tt.want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllForProject(tt.args.ctx, tt.args.project, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_GetAllForIssue(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository
		issueRepo func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository
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
		want    []*model.Issue
		wantErr error
	}{
		{
			name: "get uncached issues",
			fields: fields{
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issues,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForIssue", ctx, issue, offset, limit).Return(issues, nil)
					return repo
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.Issue{
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issues, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _, _ int, _ []*model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx:   context.Background(),
				issue: model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.Issue{
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, _ []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, _ []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForIssue", ctx, issue, offset, limit).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, _ []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _, _ int, _ []*model.Issue) repository.IssueRepository {
					return new(mock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issues,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetAllForIssue", ctx, issue, offset, limit).Return(issues, nil)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit, tt.want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit, tt.want),
			}
			got, err := r.GetAllForIssue(tt.args.ctx, tt.args.issue, tt.args.offset, tt.args.limit)
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCachedIssueRepository_AddWatcher(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id, watcher model.ID) *baseRepository
		issueRepo func(ctx context.Context, id, watcher model.ID) repository.IssueRepository
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(nil)
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(repository.ErrNotFound)
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
			name: "add watcher with clear cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(nil)
					return repo
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},

		{
			name: "add watcher with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add watcher with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddWatcher", ctx, id, watcher).Return(nil)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.watcher),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, tt.args.watcher),
			}
			err := r.AddWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetWatchers(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository
		issueRepo func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.User
		wantErr error
	}{
		{
			name: "get issue watchers",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: watchers,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetWatchers", ctx, id).Return(watchers, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, _ []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, _ []*model.User) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetWatchers", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(watchers, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ []*model.User) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: watchers,
					}).Return(repository.ErrCacheWrite)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetWatchers", ctx, id).Return(watchers, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.User{
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
				cacheRepo: func(ctx context.Context, id model.ID, _ []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, repository.ErrCacheRead)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ []*model.User) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.User{
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, tt.want),
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
		cacheRepo func(ctx context.Context, id, watcher model.ID) *baseRepository
		issueRepo func(ctx context.Context, id, watcher model.ID) repository.IssueRepository
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(nil)
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(repository.ErrNotFound)
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
			name: "remove issue watcher with clear cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(nil)
					return repo
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
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "remove issue watcher with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id, _ model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveWatcher", ctx, id, watcher).Return(nil)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.args.watcher),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, tt.args.watcher),
			}
			err := r.RemoveWatcher(tt.args.ctx, tt.args.id, tt.args.watcher)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_AddRelation(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, relation *model.IssueRelation) *baseRepository
		issueRepo func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository
	}
	type args struct {
		ctx      context.Context
		relation *model.IssueRelation
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
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "add issue relation non-issue relation",
			fields: fields{
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Target.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Target.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeDocument),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
		},
		{
			name: "add issue relation with deletion error",
			fields: fields{
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
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
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add issue relation with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Source.String(), "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add issue relation with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "add issue relation with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, relation *model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), relation.Source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", relation.Source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")
					allForProjectKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					allForProjectKeyResult := new(redis.StringSliceCmd)
					allForProjectKeyResult.SetVal([]string{allForProjectKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, relation *model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("AddRelation", ctx, relation).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				relation: &model.IssueRelation{
					Source: model.MustNewID(model.ResourceTypeIssue),
					Target: model.MustNewID(model.ResourceTypeIssue),
					Kind:   model.IssueRelationKindBlocks,
				},
			},
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.relation),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.relation),
			}
			err := r.AddRelation(tt.args.ctx, tt.args.relation)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_GetRelations(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, relations []*model.IssueRelation) *baseRepository
		issueRepo func(ctx context.Context, id model.ID, relations []*model.IssueRelation) repository.IssueRepository
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.IssueRelation
		wantErr error
	}{
		{
			name: "get issue relations",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, relations []*model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: relations,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, relations []*model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetRelations", ctx, id).Return(relations, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.IssueRelation{
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
				cacheRepo: func(ctx context.Context, id model.ID, _ []*model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, _ []*model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetRelations", ctx, id).Return(nil, repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID, relations []*model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(relations, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ []*model.IssueRelation) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.IssueRelation{
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
				cacheRepo: func(ctx context.Context, id model.ID, relations []*model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: relations,
					}).Return(repository.ErrCacheWrite)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, relations []*model.IssueRelation) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("GetRelations", ctx, id).Return(relations, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.IssueRelation{
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
				cacheRepo: func(ctx context.Context, id model.ID, _ []*model.IssueRelation) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, repository.ErrCacheRead)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(_ context.Context, _ model.ID, _ []*model.IssueRelation) repository.IssueRepository {
					return new(mock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: []*model.IssueRelation{
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, tt.want),
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
		cacheRepo func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) *baseRepository
		issueRepo func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository
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
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
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
				cacheRepo: func(ctx context.Context, _, target model.ID, _ model.IssueRelationKind) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
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
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
					return repo
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
		{
			name: "remove issue relation with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
					return repo
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
		{
			name: "remove issue relation with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), source.String())
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", source.String(), "*")

					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
					return repo
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
		{
			name: "remove issue relation with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, source, _ model.ID, _ model.IssueRelationKind) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("RemoveRelation", ctx, source, target, kind).Return(nil)
					return repo
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind),
			}
			err := r.RemoveRelation(tt.args.ctx, tt.args.source, tt.args.target, tt.args.kind)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedIssueRepository_Update(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository
		issueRepo func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository
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
		want    *model.Issue
		wantErr error
	}{
		{
			name: "update issue",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, forIssueKey).Return(forIssueKeyCmd, nil)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, forIssueKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Update", ctx, id, patch).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				patch: map[string]any{
					"title":       "new title",
					"description": "new description",
				},
			},
			want: &model.Issue{
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
				cacheRepo: func(_ context.Context, _ model.ID, _ *model.Issue) *baseRepository {
					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					return &baseRepository{
						db:     db,
						cache:  new(mock.CacheRepository),
						tracer: new(mock.Tracer),
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, _ *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Update", ctx, id, patch).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				patch: map[string]any{
					"title":       "new title",
					"description": "new description",
				},
			},
			want: &model.Issue{
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
			wantErr: repository.ErrNotFound,
		},
		{
			name: "update issue set cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					dbClient := new(mock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(assert.AnError)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Update", ctx, id, patch).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				patch: map[string]any{
					"title":       "new title",
					"description": "new description",
				},
			},
			want: &model.Issue{
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
			wantErr: repository.ErrCacheWrite,
		},
		{
			name: "update issue delete for issue to cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, forIssueKey).Return(forIssueKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, forIssueKey).Return(assert.AnError)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Update", ctx, id, patch).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				patch: map[string]any{
					"title":       "new title",
					"description": "new description",
				},
			},
			want: &model.Issue{
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
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "update issue with delete projects cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					projectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					forIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.Parent.String(), "*")

					projectsKeyCmd := new(redis.StringSliceCmd)
					projectsKeyCmd.SetVal([]string{projectsKey})

					forIssueKeyCmd := new(redis.StringSliceCmd)
					forIssueKeyCmd.SetVal([]string{forIssueKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, forIssueKey).Return(forIssueKeyCmd, nil)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyCmd, nil)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(new(redis.StatusCmd))

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)
					cacheRepo.On("Delete", ctx, forIssueKey).Return(nil)
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Update", ctx, id, patch).Return(issue, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
				patch: map[string]any{
					"title":       "new title",
					"description": "new description",
				},
			},
			want: &model.Issue{
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
			wantErr: repository.ErrCacheDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id, tt.want),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id, tt.args.patch, tt.want),
			}
			got, err := r.Update(tt.args.ctx, tt.args.id, tt.args.patch)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCachedIssueRepository_Delete(t *testing.T) {
	type fields struct {
		cacheRepo func(ctx context.Context, id model.ID) *baseRepository
		issueRepo func(ctx context.Context, id model.ID) repository.IssueRepository
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(repository.ErrNotFound)
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
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(mock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete issue with clear watchers cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete issue with clear relations cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					relationsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					relationsKeyResult := new(redis.StringSliceCmd)
					relationsKeyResult.SetVal([]string{relationsKey})

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete issue with clear for issue cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete issue with clear for project cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			wantErr: repository.ErrCacheDelete,
		},
		{
			name: "delete issue with clear projects cache error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID) *baseRepository {
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

					dbClient := new(mock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(mock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(mock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(mock.IssueRepository)
					repo.On("Delete", ctx, id).Return(nil)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &CachedIssueRepository{
				cacheRepo: tt.fields.cacheRepo(tt.args.ctx, tt.args.id),
				issueRepo: tt.fields.issueRepo(tt.args.ctx, tt.args.id),
			}
			err := r.Delete(tt.args.ctx, tt.args.id)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
