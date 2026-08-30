package async

import (
	"context"

	"github.com/hibiken/asynq"
)

// CustomFieldReconcileTaskHandler commits or aborts stale hybrid operations.
type CustomFieldReconcileTaskHandler struct {
	*baseTaskHandler
}

func (h *CustomFieldReconcileTaskHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	ctx, span := h.tracer.Start(ctx, "transport.asynq.CustomFieldReconcileTaskHandler/ProcessTask")
	defer span.End()
	return h.customFieldService.ReconcilePending(ctx)
}

func NewCustomFieldReconcileTaskHandler(opts ...TaskHandlerOption) (*CustomFieldReconcileTaskHandler, error) {
	h, err := newBaseTaskHandler(opts...)
	if err != nil {
		return nil, err
	}
	if h.customFieldService == nil {
		return nil, ErrNoCustomFieldService
	}
	return &CustomFieldReconcileTaskHandler{h}, nil
}
