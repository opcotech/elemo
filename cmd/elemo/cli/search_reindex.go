package cli

import (
	"context"

	"log/slog"

	"github.com/spf13/cobra"

	"github.com/opcotech/elemo/internal/queue"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

var (
	searchReindexAsync       bool
	searchReindexDeleteAll   bool
	searchReindexBatchSize   int
	searchReindexConcurrency int
)

func searchReindexOptions() service.SearchReindexOptions {
	opts := service.SearchReindexOptions{
		BatchSize:   searchReindexBatchSize,
		Concurrency: searchReindexConcurrency,
		DeleteAll:   searchReindexDeleteAll,
	}
	if opts.BatchSize == 0 {
		opts.BatchSize = cfg.Search.ReindexBatchSize
	}
	if opts.Concurrency == 0 {
		opts.Concurrency = cfg.Search.ReindexConcurrency
	}
	return opts
}

func initSearchService() (*repository.Neo4jDatabase, service.SearchService, error) {
	graphDB, err := initGraphDatabase()
	if err != nil {
		return nil, nil, err
	}

	_, searchRepo, err := initSearchDatabase()
	if err != nil {
		return nil, nil, err
	}

	permissionRepo, err := repository.NewNeo4jPermissionRepository(
		repository.WithNeo4jDatabase(graphDB),
		repository.WithNeo4jRepositoryLogger(logger.Named("permission_repository")),
		repository.WithNeo4jRepositoryTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	permissionService, err := service.NewPermissionService(
		permissionRepo,
		service.WithLogger(logger.Named("permission_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	searchService, err := service.NewSearchService(
		searchRepo,
		service.WithPermissionService(permissionService),
		service.WithLogger(logger.Named("search_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	return graphDB, searchService, nil
}

// searchReindexCmd backfills the Meilisearch index from Neo4j.
var searchReindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the search index from Neo4j",
	Long: `Walk the graph database and upsert searchable documents into Meilisearch.

By default the command runs in-process. Use --async to enqueue work for
background workers instead. --delete-all wipes the live index before rebuild,
which makes search incomplete until the run finishes and can leave a partial
index if a later batch fails.`,
	Run: func(_ *cobra.Command, _ []string) {
		initTracer("cli-search-reindex")

		opts := searchReindexOptions()

		if searchReindexAsync {
			messageQueue, err := queue.NewClient(
				queue.WithClientConfig(&cfg.Worker),
				queue.WithClientLogger(logger.Named("message_queue")),
				queue.WithClientTracer(tracer),
			)
			if err != nil {
				logger.Fatal(context.Background(), "failed to initialize message queue", slog.Any("error", err))
			}
			defer func() {
				if closeErr := messageQueue.Close(context.Background()); closeErr != nil {
					logger.Error(context.Background(), "failed to close message queue", slog.Any("error", closeErr))
				}
			}()

			task, err := queue.NewSearchReindexTask(queue.SearchReindexTaskPayload{
				DeleteAll: opts.DeleteAll,
				BatchSize: opts.BatchSize,
			})
			if err != nil {
				logger.Fatal(context.Background(), "failed to create search reindex task", slog.Any("error", err))
			}
			if _, err := messageQueue.Enqueue(context.Background(), task); err != nil {
				logger.Fatal(context.Background(), "failed to enqueue search reindex", slog.Any("error", err))
			}
			logger.Info(context.Background(), "search reindex enqueued")
			return
		}

		graphDB, searchService, err := initSearchService()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search service", slog.Any("error", err))
		}

		logger.Info(context.Background(), "reindexing search documents")
		if err := searchService.Reindex(context.Background(), service.SearchReindexSources{
			DB: graphDB,
		}, opts); err != nil {
			logger.Fatal(context.Background(), "failed to reindex search documents", slog.Any("error", err))
		}
		logger.Info(context.Background(), "search reindex complete")
	},
}

func init() {
	searchReindexCmd.Flags().BoolVar(&searchReindexAsync, "async", false, "enqueue reindex for background workers and return")
	searchReindexCmd.Flags().BoolVar(&searchReindexDeleteAll, "delete-all", false, "wipe the search index before rebuilding")
	searchReindexCmd.Flags().IntVar(&searchReindexBatchSize, "batch-size", 0, "documents per upsert (0 uses config/default)")
	searchReindexCmd.Flags().IntVar(&searchReindexConcurrency, "concurrency", 0, "parallel upsert workers (0 uses config/default)")
	searchCmd.AddCommand(searchReindexCmd)
}
