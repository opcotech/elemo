package async

import (
	"context"
	"errors"

	"github.com/goccy/go-json"
	"github.com/hibiken/asynq"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

// SearchIndexTaskHandler indexes one resource from current Neo4j state.
type SearchIndexTaskHandler struct {
	*baseTaskHandler
}

func (h *SearchIndexTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.SearchIndexTaskHandler/ProcessTask")
	defer span.End()

	var payload queue.SearchIndexTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	id, err := model.ParseCompositeID(payload.ResourceID)
	if err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	return h.searchService.IndexIDs(ctx, h.graphDB, id)
}

func NewSearchIndexTaskHandler(opts ...TaskHandlerOption) (*SearchIndexTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}
	if h.searchService == nil {
		return nil, ErrNoSearchService
	}
	if h.graphDB == nil {
		return nil, ErrNoGraphDatabase
	}
	return &SearchIndexTaskHandler{h}, nil
}

type searchableIDLister func(ctx context.Context, resourceType model.ResourceType) ([]model.ID, error)

// SearchReindexTaskHandler fans out batch index tasks for a full rebuild.
type SearchReindexTaskHandler struct {
	*baseTaskHandler
	listIDs searchableIDLister
}

func (h *SearchReindexTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.SearchReindexTaskHandler/ProcessTask")
	defer span.End()

	var payload queue.SearchReindexTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	if payload.DeleteAll {
		if err := h.searchService.DeleteAll(ctx); err != nil {
			return err
		}
	}

	batchSize := payload.BatchSize
	if batchSize <= 0 {
		batchSize = h.reindexBatchSize
	}
	if batchSize <= 0 {
		batchSize = service.DefaultSearchReindexBatchSize
	}

	types := []model.ResourceType{
		model.ResourceTypeOrganization,
		model.ResourceTypeNamespace,
		model.ResourceTypeProject,
		model.ResourceTypeIssue,
		model.ResourceTypeDocument,
	}
	for _, resourceType := range types {
		ids, err := h.listIDs(ctx, resourceType)
		if err != nil {
			return err
		}
		for _, chunk := range service.ChunkSearchableIDs(ids, batchSize) {
			batchTask, err := queue.NewSearchReindexBatchTask(resourceType, chunk)
			if err != nil {
				return err
			}
			if _, err := h.queueClient.Enqueue(ctx, batchTask); err != nil {
				if errors.Is(err, asynq.ErrTaskIDConflict) {
					continue
				}
				return err
			}
		}
	}
	return nil
}

func NewSearchReindexTaskHandler(opts ...TaskHandlerOption) (*SearchReindexTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}
	if h.searchService == nil {
		return nil, ErrNoSearchService
	}
	if h.graphDB == nil {
		return nil, ErrNoGraphDatabase
	}
	if h.queueClient == nil {
		return nil, ErrNoQueueClient
	}
	return &SearchReindexTaskHandler{
		baseTaskHandler: h,
		listIDs: func(ctx context.Context, resourceType model.ResourceType) ([]model.ID, error) {
			return repository.ListSearchableIDs(ctx, h.graphDB, resourceType)
		},
	}, nil
}

// SearchReindexBatchTaskHandler indexes one batch of resources.
type SearchReindexBatchTaskHandler struct {
	*baseTaskHandler
}

func (h *SearchReindexBatchTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.SearchReindexBatchTaskHandler/ProcessTask")
	defer span.End()

	var payload queue.SearchReindexBatchTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
	}

	ids := make([]model.ID, 0, len(payload.IDs))
	for _, raw := range payload.IDs {
		id, err := model.ParseCompositeID(raw)
		if err != nil {
			return errors.Join(ErrTaskPayloadUnmarshal, err, asynq.SkipRetry)
		}
		ids = append(ids, id)
	}

	return h.searchService.IndexIDs(ctx, h.graphDB, ids...)
}

func NewSearchReindexBatchTaskHandler(opts ...TaskHandlerOption) (*SearchReindexBatchTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}
	if h.searchService == nil {
		return nil, ErrNoSearchService
	}
	if h.graphDB == nil {
		return nil, ErrNoGraphDatabase
	}
	return &SearchReindexBatchTaskHandler{h}, nil
}
