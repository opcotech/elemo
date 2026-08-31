package service

import (
	"context"
	"errors"
	"slices"

	"github.com/hibiken/asynq"
	"golang.org/x/sync/errgroup"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/event"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
)

// SearchTaskEnqueuer schedules search-related background tasks.
type SearchTaskEnqueuer interface {
	Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

func (o SearchReindexOptions) normalized() SearchReindexOptions {
	if o.BatchSize <= 0 {
		o.BatchSize = DefaultSearchReindexBatchSize
	}
	if o.BatchSize > repository.MaxPageSize {
		o.BatchSize = repository.MaxPageSize
	}
	if o.Concurrency <= 0 {
		o.Concurrency = DefaultSearchReindexConcurrency
	}
	return o
}

func (s *searchService) EnqueueIndex(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/EnqueueIndex")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if s.searchTaskEnqueuer == nil {
		return errors.Join(ErrSearchIndex, ErrNoSearchTaskEnqueuer)
	}

	task, err := queue.NewSearchIndexTask(id)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if _, err := s.searchTaskEnqueuer.Enqueue(ctx, task); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	return nil
}

func (s *searchService) DeleteAll(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/DeleteAll")
	defer span.End()

	if err := s.searchRepo.DeleteAll(ctx); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

func (s *searchService) IndexIDs(ctx context.Context, db *repository.Neo4jDatabase, ids ...model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/IndexIDs")
	defer span.End()

	if db == nil {
		return errors.Join(ErrSearchIndex, repository.ErrNoDriver)
	}
	if len(ids) == 0 {
		return nil
	}

	byType := make(map[model.ResourceType][]model.ID)
	for _, id := range ids {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrSearchIndex, err)
		}
		byType[id.Type] = append(byType[id.Type], id)
	}

	docs := make([]repository.SearchDocument, 0, len(ids))
	for resourceType, typedIDs := range byType {
		records, err := s.listSearchableByIDs(ctx, db, resourceType, typedIDs)
		if err != nil {
			return errors.Join(ErrSearchIndex, err)
		}
		for _, rec := range records {
			docs = append(docs, searchDocumentFromRecord(rec))
		}
	}
	if len(docs) == 0 {
		return nil
	}
	if err := s.searchRepo.Upsert(ctx, docs...); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	return nil
}

func (s *searchService) Reindex(ctx context.Context, sources SearchReindexSources, opts SearchReindexOptions) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/Reindex")
	defer span.End()

	if sources.DB == nil {
		return errors.Join(ErrSearchReindex, repository.ErrNoDriver)
	}

	opts = opts.normalized()
	if opts.DeleteAll {
		if err := s.searchRepo.DeleteAll(ctx); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(opts.Concurrency)

	for _, resourceType := range searchableResourceTypes {
		after := ""
		for {
			if err := gctx.Err(); err != nil {
				if waitErr := g.Wait(); waitErr != nil {
					return waitErr
				}
				return errors.Join(ErrSearchReindex, err)
			}

			records, err := s.listSearchableRecords(gctx, sources.DB, resourceType, after, opts.BatchSize)
			if err != nil {
				_ = g.Wait()
				return errors.Join(ErrSearchReindex, err)
			}
			if len(records) == 0 {
				break
			}

			batch := slices.Clone(records)
			kind := resourceType
			cursor := after
			g.Go(func() error {
				docs := make([]repository.SearchDocument, len(batch))
				for i, rec := range batch {
					docs[i] = searchDocumentFromRecord(rec)
				}
				s.logger.Info(gctx, "reindexing search documents",
					log.WithKind(kind.String()),
					log.WithLimit(len(docs)),
					log.WithValue(cursor),
				)
				if err := s.searchRepo.Upsert(gctx, docs...); err != nil {
					return errors.Join(ErrSearchReindex, err)
				}
				return nil
			})

			after = records[len(records)-1].ID.String()
			if len(records) < opts.BatchSize {
				break
			}
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func enqueueSearchIndex(ctx context.Context, logger log.Logger, searchService SearchService, id model.ID) {
	if err := searchService.EnqueueIndex(ctx, id); err != nil {
		logger.Warn(ctx, "failed to enqueue search index",
			log.WithError(err),
			log.WithValue(id.Composite()),
		)
	}
}

func publishDomainEvent(ctx context.Context, bus EventPublisher, logger log.Logger, evt event.Event) {
	if bus == nil {
		return
	}
	if err := bus.Publish(ctx, evt); err != nil {
		logger.Warn(ctx, "failed to publish domain event",
			log.WithError(err),
			log.WithValue(string(evt.Type)),
		)
	}
}

func ChunkSearchableIDs(ids []model.ID, size int) [][]model.ID {
	if size <= 0 {
		size = DefaultSearchReindexBatchSize
	}
	if len(ids) == 0 {
		return [][]model.ID{}
	}
	chunks := make([][]model.ID, 0, (len(ids)+size-1)/size)
	for i := 0; i < len(ids); i += size {
		end := min(i+size, len(ids))
		chunks = append(chunks, ids[i:end])
	}
	return chunks
}
