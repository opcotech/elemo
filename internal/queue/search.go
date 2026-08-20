package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/model"
)

const (
	SearchIndexTaskTimeout        = 30 * time.Second
	SearchReindexTaskTimeout      = 5 * time.Minute
	SearchReindexBatchTaskTimeout = 2 * time.Minute
	searchReindexTaskID           = "search:reindex"
)

// SearchIndexTaskPayload identifies one resource to index from current graph state.
type SearchIndexTaskPayload struct {
	ResourceID string `json:"resource_id"`
}

// SearchReindexTaskPayload starts a full search rebuild.
type SearchReindexTaskPayload struct {
	DeleteAll bool `json:"delete_all"`
	BatchSize int  `json:"batch_size"`
}

// SearchReindexBatchTaskPayload indexes one page of resources of a single type.
type SearchReindexBatchTaskPayload struct {
	ResourceType string   `json:"resource_type"`
	IDs          []string `json:"ids"`
}

// NewSearchIndexTask creates a write-through index task for one resource.
func NewSearchIndexTask(id model.ID) (*asynq.Task, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(SearchIndexTaskPayload{ResourceID: id.Composite()})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TaskTypeSearchIndex.String(),
		payload,
		asynq.Timeout(SearchIndexTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
		asynq.Queue(MessageQueueDefaultPriority),
	), nil
}

// NewSearchReindexTask creates a unique coordinator task for a full rebuild.
func NewSearchReindexTask(payload SearchReindexTaskPayload) (*asynq.Task, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TaskTypeSearchReindex.String(),
		raw,
		asynq.TaskID(searchReindexTaskID),
		asynq.Timeout(SearchReindexTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
		asynq.Queue(MessageQueueLowPriority),
	), nil
}

// NewSearchReindexBatchTask creates a batch index task.
func NewSearchReindexBatchTask(resourceType model.ResourceType, ids []model.ID) (*asynq.Task, error) {
	if !resourceType.IsAResourceType() {
		return nil, model.ErrInvalidResourceType
	}

	rawIDs := make([]string, len(ids))
	for i, id := range ids {
		rawIDs[i] = id.Composite()
	}

	payload, err := json.Marshal(SearchReindexBatchTaskPayload{
		ResourceType: resourceType.String(),
		IDs:          rawIDs,
	})
	if err != nil {
		return nil, err
	}

	taskID := searchReindexTaskID + "_batch:" + resourceType.String()
	if len(ids) > 0 {
		taskID += ":" + ids[0].String() + ":" + ids[len(ids)-1].String()
	}

	return asynq.NewTask(
		TaskTypeSearchReindexBatch.String(),
		payload,
		asynq.TaskID(taskID),
		asynq.Timeout(SearchReindexBatchTaskTimeout),
		asynq.Retention(DefaultTaskRetention),
		asynq.Queue(MessageQueueLowPriority),
	), nil
}
