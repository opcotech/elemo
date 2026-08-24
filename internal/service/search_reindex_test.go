package service_test

import (
	"context"
	"testing"

	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
)

type stubSearchEnqueuer struct {
	task *asynq.Task
	err  error
}

func (s *stubSearchEnqueuer) Enqueue(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	s.task = task
	return &asynq.TaskInfo{}, s.err
}

func TestSearchService_EnqueueIndex(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("requires enqueuer", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), mockrepo.NewMockSearchRepository(ctrl))
		err := svc.EnqueueIndex(context.Background(), issueID)
		assert.ErrorIs(t, err, service.ErrNoSearchTaskEnqueuer)
	})

	t.Run("enqueues resource id task", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		enqueuer := &stubSearchEnqueuer{}
		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), mockrepo.NewMockSearchRepository(ctrl))
		service.SetSearchServiceEnqueuer(svc, enqueuer)

		err := svc.EnqueueIndex(context.Background(), issueID)
		require.NoError(t, err)
		require.NotNil(t, enqueuer.task)
		assert.Equal(t, queue.TaskTypeSearchIndex.String(), enqueuer.task.Type())
		assert.Contains(t, string(enqueuer.task.Payload()), issueID.Composite())
	})

	t.Run("preserves enqueue errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		enqueuer := &stubSearchEnqueuer{err: assert.AnError}
		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), mockrepo.NewMockSearchRepository(ctrl))
		service.SetSearchServiceEnqueuer(svc, enqueuer)

		err := svc.EnqueueIndex(context.Background(), issueID)
		assert.ErrorIs(t, err, service.ErrSearchIndex)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestSearchService_IndexIDs(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	projectID := model.MustNewID(model.ResourceTypeProject)
	ctrl := gomock.NewController(t)
	repo := mockrepo.NewMockSearchRepository(ctrl)
	svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
	service.SetSearchServiceListByIDs(svc, func(
		_ context.Context,
		_ *repository.Neo4jDatabase,
		resourceType model.ResourceType,
		ids []model.ID,
	) ([]repository.SearchableRecord, error) {
		assert.Equal(t, model.ResourceTypeIssue, resourceType)
		assert.Equal(t, []model.ID{issueID}, ids)
		return []repository.SearchableRecord{{
			ID:       issueID,
			Title:    "Bug",
			Ancestry: []model.ID{issueID, projectID},
		}}, nil
	})
	repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, docs ...repository.SearchDocument) error {
			require.Len(t, docs, 1)
			assert.Equal(t, issueID.SearchKey(), docs[0].ID)
			return nil
		},
	)

	err := svc.IndexIDs(context.Background(), &repository.Neo4jDatabase{}, issueID)
	require.NoError(t, err)
}

func TestSearchService_IndexIDs_MissingResourceNoops(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mockrepo.NewMockSearchRepository(ctrl)
	svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
	service.SetSearchServiceListByIDs(svc, func(
		_ context.Context,
		_ *repository.Neo4jDatabase,
		_ model.ResourceType,
		_ []model.ID,
	) ([]repository.SearchableRecord, error) {
		return []repository.SearchableRecord{}, nil
	})

	err := svc.IndexIDs(context.Background(), &repository.Neo4jDatabase{}, model.MustNewID(model.ResourceTypeIssue))
	require.NoError(t, err)
}

func TestChunkSearchableIDs(t *testing.T) {
	t.Parallel()

	ids := []model.ID{
		model.MustNewID(model.ResourceTypeIssue),
		model.MustNewID(model.ResourceTypeIssue),
		model.MustNewID(model.ResourceTypeIssue),
	}
	chunks := service.ChunkSearchableIDs(ids, 2)
	require.Len(t, chunks, 2)
	assert.Len(t, chunks[0], 2)
	assert.Len(t, chunks[1], 1)
	assert.Empty(t, service.ChunkSearchableIDs(nil, 2))
}
