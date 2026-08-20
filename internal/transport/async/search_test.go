package async

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

type stubTaskEnqueuer struct {
	tasks []*asynq.Task
}

func (s *stubTaskEnqueuer) Enqueue(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	s.tasks = append(s.tasks, task)
	return &asynq.TaskInfo{}, nil
}

func testSearchHandlerBase(ctx context.Context, ctrl *gomock.Controller, spanName string) *baseTaskHandler {
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End().Return()
	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, spanName).Return(ctx, span)
	return &baseTaskHandler{
		logger:           mock.NewMockLogger(nil),
		tracer:           tracer,
		reindexBatchSize: service.DefaultSearchReindexBatchSize,
	}
}

func TestSearchIndexTaskHandler_ProcessTask(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issueID := model.MustNewID(model.ResourceTypeIssue)
	db := &repository.Neo4jDatabase{}

	t.Run("indexes current resource", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		searchSvc := service.NewMockSearchService(ctrl)
		searchSvc.EXPECT().IndexIDs(ctx, db, issueID).Return(nil)
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchIndexTaskHandler/ProcessTask")
		base.searchService = searchSvc
		base.graphDB = db

		task, err := queue.NewSearchIndexTask(issueID)
		require.NoError(t, err)
		err = (&SearchIndexTaskHandler{baseTaskHandler: base}).ProcessTask(ctx, task)
		require.NoError(t, err)
	})

	t.Run("invalid payload skips retry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchIndexTaskHandler/ProcessTask")
		err := (&SearchIndexTaskHandler{baseTaskHandler: base}).ProcessTask(ctx, asynq.NewTask(
			queue.TaskTypeSearchIndex.String(),
			[]byte(`{`),
		))
		assert.ErrorIs(t, err, ErrTaskPayloadUnmarshal)
		assert.ErrorIs(t, err, asynq.SkipRetry)
	})
}

func TestSearchReindexTaskHandler_ProcessTask(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issueID := model.MustNewID(model.ResourceTypeIssue)
	db := &repository.Neo4jDatabase{}

	t.Run("wipes then enqueues batches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		searchSvc := service.NewMockSearchService(ctrl)
		listed := false
		searchSvc.EXPECT().DeleteAll(ctx).DoAndReturn(func(context.Context) error {
			assert.False(t, listed)
			return nil
		})
		enqueuer := &stubTaskEnqueuer{}
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchReindexTaskHandler/ProcessTask")
		base.searchService = searchSvc
		base.graphDB = db
		base.queueClient = enqueuer
		base.reindexBatchSize = 10

		handler := &SearchReindexTaskHandler{
			baseTaskHandler: base,
			listIDs: func(_ context.Context, resourceType model.ResourceType) ([]model.ID, error) {
				listed = true
				if resourceType == model.ResourceTypeIssue {
					return []model.ID{issueID}, nil
				}
				return []model.ID{}, nil
			},
		}

		task, err := queue.NewSearchReindexTask(queue.SearchReindexTaskPayload{DeleteAll: true, BatchSize: 10})
		require.NoError(t, err)
		require.NoError(t, handler.ProcessTask(ctx, task))
		require.Len(t, enqueuer.tasks, 1)
		assert.Equal(t, queue.TaskTypeSearchReindexBatch.String(), enqueuer.tasks[0].Type())
		assert.True(t, listed)
	})

	t.Run("does not wipe without delete_all", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		searchSvc := service.NewMockSearchService(ctrl)
		enqueuer := &stubTaskEnqueuer{}
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchReindexTaskHandler/ProcessTask")
		base.searchService = searchSvc
		base.graphDB = db
		base.queueClient = enqueuer

		handler := &SearchReindexTaskHandler{
			baseTaskHandler: base,
			listIDs: func(_ context.Context, resourceType model.ResourceType) ([]model.ID, error) {
				if resourceType == model.ResourceTypeIssue {
					return []model.ID{issueID}, nil
				}
				return []model.ID{}, nil
			},
		}

		task, err := queue.NewSearchReindexTask(queue.SearchReindexTaskPayload{})
		require.NoError(t, err)
		require.NoError(t, handler.ProcessTask(ctx, task))
		require.Len(t, enqueuer.tasks, 1)
	})

	t.Run("invalid payload skips retry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchReindexTaskHandler/ProcessTask")
		err := (&SearchReindexTaskHandler{baseTaskHandler: base}).ProcessTask(ctx, asynq.NewTask(
			queue.TaskTypeSearchReindex.String(),
			[]byte(`{`),
		))
		assert.ErrorIs(t, err, ErrTaskPayloadUnmarshal)
		assert.ErrorIs(t, err, asynq.SkipRetry)
	})
}

func TestSearchReindexBatchTaskHandler_ProcessTask(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	issueID := model.MustNewID(model.ResourceTypeIssue)
	db := &repository.Neo4jDatabase{}

	t.Run("indexes batch", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		searchSvc := service.NewMockSearchService(ctrl)
		searchSvc.EXPECT().IndexIDs(ctx, db, issueID).Return(nil)
		base := testSearchHandlerBase(ctx, ctrl, "transport.asynq.SearchReindexBatchTaskHandler/ProcessTask")
		base.searchService = searchSvc
		base.graphDB = db

		task, err := queue.NewSearchReindexBatchTask(model.ResourceTypeIssue, []model.ID{issueID})
		require.NoError(t, err)
		require.NoError(t, (&SearchReindexBatchTaskHandler{baseTaskHandler: base}).ProcessTask(ctx, task))
	})
}

func TestNewSearchIndexTaskHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	_, err := NewSearchIndexTaskHandler(
		WithTaskLogger(mock.NewMockLogger(nil)),
		WithTaskTracer(mock.NewMockTracer(nil)),
	)
	assert.ErrorIs(t, err, ErrNoSearchService)

	searchSvc := service.NewMockSearchService(ctrl)
	_, err = NewSearchIndexTaskHandler(
		WithTaskSearchService(searchSvc),
		WithTaskLogger(log.DefaultLogger()),
		WithTaskTracer(tracing.NoopTracer()),
	)
	assert.ErrorIs(t, err, ErrNoGraphDatabase)
}

func TestNewSearchReindexTaskHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	searchSvc := service.NewMockSearchService(ctrl)
	db := &repository.Neo4jDatabase{}

	_, err := NewSearchReindexTaskHandler(
		WithTaskSearchService(searchSvc),
		WithTaskGraphDatabase(db),
		WithTaskLogger(log.DefaultLogger()),
		WithTaskTracer(tracing.NoopTracer()),
	)
	assert.ErrorIs(t, err, ErrNoQueueClient)
}
