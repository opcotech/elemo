package async

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/queue"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

func TestNewCustomFieldReconcileTaskHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		got, err := NewCustomFieldReconcileTaskHandler(
			WithTaskLogger(mocklog.NewMockLogger(nil)),
			WithTaskTracer(mocktrace.NewMockTracer(nil)),
			WithTaskCustomFieldService(cfSvc),
		)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, cfSvc, got.customFieldService)
	})

	t.Run("missing custom field service", func(t *testing.T) {
		t.Parallel()
		_, err := NewCustomFieldReconcileTaskHandler(
			WithTaskLogger(mocklog.NewMockLogger(nil)),
			WithTaskTracer(mocktrace.NewMockTracer(nil)),
		)
		assert.ErrorIs(t, err, ErrNoCustomFieldService)
	})

	t.Run("invalid option", func(t *testing.T) {
		t.Parallel()
		_, err := NewCustomFieldReconcileTaskHandler(WithTaskLogger(nil))
		assert.ErrorIs(t, err, log.ErrNoLogger)
	})
}

func TestCustomFieldReconcileTaskHandler_ProcessTask(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("reconciles pending operations", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End()
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "transport.asynq.CustomFieldReconcileTaskHandler/ProcessTask").Return(ctx, span)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().ReconcilePending(ctx).Return(nil)

		task, err := queue.NewCustomFieldReconcileTask()
		require.NoError(t, err)

		err = (&CustomFieldReconcileTaskHandler{
			baseTaskHandler: &baseTaskHandler{
				logger:             mocklog.NewMockLogger(nil),
				tracer:             tracer,
				customFieldService: cfSvc,
			},
		}).ProcessTask(ctx, task)
		require.NoError(t, err)
	})

	t.Run("returns service error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End()
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "transport.asynq.CustomFieldReconcileTaskHandler/ProcessTask").Return(ctx, span)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().ReconcilePending(ctx).Return(assert.AnError)

		task, err := queue.NewCustomFieldReconcileTask()
		require.NoError(t, err)

		err = (&CustomFieldReconcileTaskHandler{
			baseTaskHandler: &baseTaskHandler{
				logger:             mocklog.NewMockLogger(nil),
				tracer:             tracer,
				customFieldService: cfSvc,
			},
		}).ProcessTask(ctx, task)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
