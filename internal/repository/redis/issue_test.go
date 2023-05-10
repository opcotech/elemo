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
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	testMock "github.com/opcotech/elemo/internal/testutil/mock"
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
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, parentIssueKey).Return(parentIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, parentIssueKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, parentIssueKey).Return(parentIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, projectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, parentIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, project model.ID, issue *model.Issue) *baseRepository {
					allProjectsKey := composeCacheKey(model.ResourceTypeProject.String(), "*")
					projectsKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), "*")

					allProjectsKeyResult := new(redis.StringSliceCmd)
					allProjectsKeyResult.SetVal([]string{allProjectsKey})

					projectsKeyResult := new(redis.StringSliceCmd)
					projectsKeyResult.SetVal([]string{projectsKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)
					dbClient.On("Keys", ctx, allProjectsKey).Return(allProjectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, allProjectsKey).Return(nil)
					cacheRepo.On("Delete", ctx, projectsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, issue *model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				}
			},
		},
		{
			name: "get cached issue",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issue, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.MustNewID(model.ResourceTypeIssue),
			},
			want: func(id model.ID) *model.Issue {
				return &model.Issue{
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
				}
			},
		},
		{
			name: "get uncached issue error",
			fields: fields{
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

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
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
						Value: issue,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						Value: issues,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issues, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", project.String(), offset, limit)

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
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
						Value: issues,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, project model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						Value: issues,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(issues, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issue.String(), offset, limit)

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
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
						Value: issues,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, issue model.ID, offset, limit int, issues []*model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())

					db, err := NewDatabase(
						WithClient(new(testMock.RedisClient)),
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
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
					watchersKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String(), "*")
					allForIssueKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", "*")

					watchersKeyResult := new(redis.StringSliceCmd)
					watchersKeyResult.SetVal([]string{watchersKey})

					allForIssueKeyResult := new(redis.StringSliceCmd)
					allForIssueKeyResult.SetVal([]string{allForIssueKey})

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id, watcher model.ID) *baseRepository {
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id, watcher model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						Value: watchers,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						WithClient(new(testMock.RedisClient)),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Get", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(watchers, nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
						Value: watchers,
					}).Return(repository.ErrCacheWrite)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, watchers []*model.User) *baseRepository {
					key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", id.String())

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
					cacheRepo.On("Get", ctx, key, mock.Anything).Return(nil, repository.ErrCacheRead)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, watchers []*model.User) repository.IssueRepository {
					return new(testMock.IssueRepository)
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
	t.Skip("not implemented")
}

func TestCachedIssueRepository_AddRelation(t *testing.T) {
	t.Skip("not implemented")
}

func TestCachedIssueRepository_GetRelations(t *testing.T) {
	t.Skip("not implemented")
}

func TestCachedIssueRepository_RemoveRelation(t *testing.T) {
	t.Skip("not implemented")
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

					dbClient := new(testMock.RedisClient)
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

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
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
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
				cacheRepo: func(ctx context.Context, id model.ID, issue *model.Issue) *baseRepository {
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
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
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
						Value: issue,
					}).Return(errors.New("error"))

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
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

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, forIssueKey).Return(errors.New("error"))
					cacheRepo.On("Set", &cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: issue,
					}).Return(nil)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
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

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Set", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
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
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID, patch map[string]any, issue *model.Issue) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
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
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
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
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
						WithClient(new(testMock.RedisClient)),
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
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
					cacheRepo.On("Delete", ctx, key).Return(nil)
					cacheRepo.On("Delete", ctx, watchersKey).Return(nil)
					cacheRepo.On("Delete", ctx, relationsKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForIssueKey).Return(nil)
					cacheRepo.On("Delete", ctx, allForProjectKey).Return(repository.ErrCacheDelete)

					return &baseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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

					dbClient := new(testMock.RedisClient)
					dbClient.On("Keys", ctx, watchersKey).Return(watchersKeyResult)
					dbClient.On("Keys", ctx, relationsKey).Return(relationsKeyResult)
					dbClient.On("Keys", ctx, allForIssueKey).Return(allForIssueKeyResult)
					dbClient.On("Keys", ctx, allForProjectKey).Return(allForProjectKeyResult)
					dbClient.On("Keys", ctx, projectsKey).Return(projectsKeyResult)

					db, err := NewDatabase(
						WithClient(dbClient),
					)
					require.NoError(t, err)

					span := new(testMock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(testMock.Tracer)
					tracer.On("Start", ctx, "repository.redis.baseRepository/DeletePattern", []trace.SpanStartOption(nil)).Return(ctx, span)
					tracer.On("Start", ctx, "repository.redis.baseRepository/Delete", []trace.SpanStartOption(nil)).Return(ctx, span)

					cacheRepo := new(testMock.CacheRepository)
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
						logger: new(testMock.Logger),
					}
				},
				issueRepo: func(ctx context.Context, id model.ID) repository.IssueRepository {
					repo := new(testMock.IssueRepository)
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
