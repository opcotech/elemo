package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestNewSearchIndexTask(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeIssue)
	got, err := NewSearchIndexTask(id)
	require.NoError(t, err)
	assert.Equal(t, TaskTypeSearchIndex.String(), got.Type())
	assert.Contains(t, string(got.Payload()), id.Composite())
}

func TestNewSearchReindexTask(t *testing.T) {
	t.Parallel()

	got, err := NewSearchReindexTask(SearchReindexTaskPayload{DeleteAll: true, BatchSize: 100})
	require.NoError(t, err)
	assert.Equal(t, TaskTypeSearchReindex.String(), got.Type())
	assert.Contains(t, string(got.Payload()), `"delete_all":true`)
	assert.Contains(t, string(got.Payload()), `"batch_size":100`)
}

func TestNewSearchReindexBatchTask(t *testing.T) {
	t.Parallel()

	ids := []model.ID{model.MustNewID(model.ResourceTypeIssue)}
	got, err := NewSearchReindexBatchTask(model.ResourceTypeIssue, ids)
	require.NoError(t, err)
	assert.Equal(t, TaskTypeSearchReindexBatch.String(), got.Type())
	assert.Contains(t, string(got.Payload()), ids[0].Composite())

	_, err = NewSearchReindexBatchTask(model.ResourceType(0), nil)
	assert.ErrorIs(t, err, model.ErrInvalidResourceType)
}

func TestNewSearchIndexTask_InvalidID(t *testing.T) {
	t.Parallel()

	task, err := NewSearchIndexTask(model.ID{})
	assert.Error(t, err)
	assert.Nil(t, task)
}
