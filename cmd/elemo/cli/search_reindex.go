package cli

import (
	"context"

	"log/slog"

	"github.com/spf13/cobra"

	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

// searchReindexCmd backfills the Meilisearch index from Neo4j.
var searchReindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the search index from Neo4j",
	Long: `Walk the graph database and upsert them into Meilisearch. Use this to
backfill an empty index or repair drift after a failed write-through.`,
	Run: func(_ *cobra.Command, _ []string) {
		initTracer("cli-search-reindex")

		graphDB, err := initGraphDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize graph database", slog.Any("error", err))
		}

		_, searchRepo, err := initSearchDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search database", slog.Any("error", err))
		}

		permissionRepo, err := repository.NewNeo4jPermissionRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("permission_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize permission repository", slog.Any("error", err))
		}

		organizationRepo, err := repository.NewNeo4jOrganizationRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("organization_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize organization repository", slog.Any("error", err))
		}

		namespaceRepo, err := repository.NewNeo4jNamespaceRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("namespace_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize namespace repository", slog.Any("error", err))
		}

		projectRepo, err := repository.NewNeo4jProjectRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("project_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize project repository", slog.Any("error", err))
		}

		issueRepo, err := repository.NewNeo4jIssueRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("issue_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize issue repository", slog.Any("error", err))
		}

		documentRepo, err := repository.NewNeo4jDocumentRepository(
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named("document_repository")),
			repository.WithNeo4jRepositoryTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize document repository", slog.Any("error", err))
		}

		permissionService, err := service.NewPermissionService(
			permissionRepo,
			service.WithLogger(logger.Named("permission_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize permission service", slog.Any("error", err))
		}

		searchService, err := service.NewSearchService(
			searchRepo,
			service.WithPermissionService(permissionService),
			service.WithLogger(logger.Named("search_service")),
			service.WithTracer(tracer),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize search service", slog.Any("error", err))
		}

		logger.Info(context.Background(), "reindexing search documents")
		if err := searchService.Reindex(context.Background(), service.SearchReindexSources{
			DB:           graphDB,
			Organization: organizationRepo,
			Namespace:    namespaceRepo,
			Project:      projectRepo,
			Issue:        issueRepo,
			Document:     documentRepo,
		}); err != nil {
			logger.Fatal(context.Background(), "failed to reindex search documents", slog.Any("error", err))
		}
		logger.Info(context.Background(), "search reindex complete")
	},
}

func init() {
	searchCmd.AddCommand(searchReindexCmd)
}
