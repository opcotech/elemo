package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/metrics"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestCachedIssueRepository_ListForProject_metrics(t *testing.T) {
	page := Page[*PartialIssue]{
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
	}
	query := IssueListQuery{
		ProjectID:  model.MustNewID(model.ResourceTypeProject),
		Page:       CursorPage{Size: 10},
		Projection: IssueListForProjectProjection(),
	}
	ctx := context.Background()

	t.Run("miss", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		baseKey, err := issueListForProjectCacheKey(query)
		require.NoError(t, err)
		key := composeCacheKey(baseKey, "g", int64(0), "ae", int64(0), "pe", int64(0))

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)

		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(5)

		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, issueListProjectGenKey(query.ProjectID), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Set(gomock.Any()).Return(nil)

		repo := NewMockIssueRepository(ctrl)
		repo.EXPECT().ListForProject(ctx, query).Return(page, nil)

		r := &RedisCachedIssueRepository{
			cacheRepo: &redisBaseRepository{
				db:     db,
				cache:  cacheRepo,
				tracer: tracer,
				logger: mock.NewMockLogger(ctrl),
			},
			issueRepo: repo,
		}

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

		baseKey, err := issueListForProjectCacheKey(query)
		require.NoError(t, err)
		key := composeCacheKey(baseKey, "g", int64(0), "ae", int64(0), "pe", int64(0))

		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)

		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).Times(4)

		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(4)

		cacheRepo := mock.NewCacheBackend(ctrl)
		cacheRepo.EXPECT().Get(ctx, issueListAuthzEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, issueListProjectionEpochKey(), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, issueListProjectGenKey(query.ProjectID), gomock.Any()).Return(cache.ErrCacheMiss)
		cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
			*(dst.(*Page[*PartialIssue])) = page
		}).Return(nil)

		r := &RedisCachedIssueRepository{
			cacheRepo: &redisBaseRepository{
				db:     db,
				cache:  cacheRepo,
				tracer: tracer,
				logger: mock.NewMockLogger(ctrl),
			},
			issueRepo: NewMockIssueRepository(ctrl),
		}

		before := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultHit))
		got, err := r.ListForProject(ctx, query)
		require.NoError(t, err)
		require.Equal(t, page, got)
		after := testutil.ToFloat64(metrics.IssueListCacheRequests.WithLabelValues(metrics.IssueListScopeProject, metrics.ResultHit))
		require.GreaterOrEqual(t, after, before+1)
	})
}
