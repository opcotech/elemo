package async

import (
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

// TaskHandlerOption is a function that can be used to configure a task handler.
type TaskHandlerOption func(*baseTaskHandler) error

// WithTaskEmailService sets the email service for the worker.
func WithTaskEmailService(emailService service.EmailService) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if emailService == nil {
			return ErrNoEmailService
		}

		t.emailService = emailService
		return nil
	}
}

// WithTaskSearchService sets the search service for the worker.
func WithTaskSearchService(searchService service.SearchService) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if searchService == nil {
			return ErrNoSearchService
		}

		t.searchService = searchService
		return nil
	}
}

// WithTaskCustomFieldService sets the custom-field service for reconcile tasks.
func WithTaskCustomFieldService(customFieldService service.CustomFieldService) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if customFieldService == nil {
			return ErrNoCustomFieldService
		}
		t.customFieldService = customFieldService
		return nil
	}
}

// WithTaskGraphDatabase sets the graph database for search tasks.
func WithTaskGraphDatabase(db *repository.Neo4jDatabase) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if db == nil {
			return ErrNoGraphDatabase
		}

		t.graphDB = db
		return nil
	}
}

// WithTaskQueueClient sets the queue client used to fan out reindex batches.
func WithTaskQueueClient(client service.SearchTaskEnqueuer) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if client == nil {
			return ErrNoQueueClient
		}

		t.queueClient = client
		return nil
	}
}

// WithTaskReindexBatchSize sets the default batch size for search reindex.
func WithTaskReindexBatchSize(size int) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		t.reindexBatchSize = size
		return nil
	}
}

// WithTaskLogger sets the logger for the task handler.
func WithTaskLogger(logger log.Logger) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		t.logger = logger

		return nil
	}
}

// WithTaskTracer sets the tracer for the task handler.
func WithTaskTracer(tracer tracing.Tracer) TaskHandlerOption {
	return func(t *baseTaskHandler) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		t.tracer = tracer

		return nil
	}
}

// baseTaskHandler serves as the base type for all task handlers.
type baseTaskHandler struct {
	logger log.Logger
	tracer tracing.Tracer

	emailService       service.EmailService
	searchService      service.SearchService
	customFieldService service.CustomFieldService
	graphDB            *repository.Neo4jDatabase
	queueClient        service.SearchTaskEnqueuer
	reindexBatchSize   int
}

// newBaseTaskHandler creates a new base task handler.
func newBaseTaskHandler(opts ...TaskHandlerOption) (*baseTaskHandler, error) {
	t := &baseTaskHandler{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}
