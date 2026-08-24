package repository_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/go-redis/cache/v9"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/metrics"
)

func TestCachedIssueRepository_ListForProject_metrics(t *testing.T) {
	page := repository.Page[*repository.PartialIssue]{
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
	}
	query := repository.IssueListQuery{
		ProjectID:  model.MustNewID(model.ResourceTypeProject),
		Page:       repository.CursorPage{Size: 10},
		Projection: repository.IssueListForProjectProjection(),
	}
	ctx := context.Background()

	t.Run("miss", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		baseKey := mustPlanCacheKey(t, query, model.ResourceTypeIssue.String(), "ListForProject", query.ProjectID.String())
		key := baseKey + ":g:0:ae:0:pe:0"

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
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

		repo := mockrepo.NewMockIssueRepository(ctrl)
		repo.EXPECT().ListForProject(ctx, query).Return(page, nil)

		r := func() *repository.RedisCachedIssueRepository {
			r, err := repository.NewCachedIssueRepository(
				repo,
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

		before := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultMiss))
		got, err := r.ListForProject(ctx, query)
		require.NoError(t, err)
		require.Equal(t, page, got)
		after := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultMiss))
		require.GreaterOrEqual(t, after, before+1)
	})

	t.Run("hit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		baseKey := mustPlanCacheKey(t, query, model.ResourceTypeIssue.String(), "ListForProject", query.ProjectID.String())
		key := baseKey + ":g:0:ae:0:pe:0"

		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
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

		before := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultHit))
		got, err := r.ListForProject(ctx, query)
		require.NoError(t, err)
		require.Equal(t, page, got)
		after := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultHit))
		require.GreaterOrEqual(t, after, before+1)
	})
}
