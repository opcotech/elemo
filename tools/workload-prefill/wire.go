package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/opcotech/elemo/assets/keys"
	"github.com/opcotech/elemo/internal/config"
	elemoLicense "github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

const envPrefix = "ELEMO"

type deps struct {
	cfg    *config.Config
	logger log.Logger
	tracer tracing.Tracer

	graphDB    *repository.Neo4jDatabase
	relDB      *repository.PGDatabase
	cacheDB    *repository.RedisDatabase
	searchRepo *repository.MeilisearchSearchRepository

	users       service.UserService
	orgs        service.OrganizationService
	teams       service.TeamService
	roles       service.RoleService
	namespaces  service.NamespaceService
	projects    service.ProjectService
	issues      service.IssueService
	documents   service.DocumentService
	permissions service.PermissionService
	search      service.SearchService
}

func (d *deps) close(ctx context.Context) error {
	var errs []error
	if d.graphDB != nil {
		errs = append(errs, d.graphDB.Close(ctx))
	}
	if d.relDB != nil {
		errs = append(errs, d.relDB.Close())
	}
	if d.cacheDB != nil {
		errs = append(errs, d.cacheDB.Close())
	}
	return errors.Join(errs...)
}

func loadConfig(path string) (*config.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

func parseLicense(ctx context.Context, logger log.Logger, licenseConf *config.LicenseConfig) (*elemoLicense.License, error) {
	if licenseConf == nil {
		return nil, elemoLicense.ErrNoLicense
	}

	data, err := os.ReadFile(licenseConf.File)
	if err != nil {
		return nil, fmt.Errorf("read license: %w", err)
	}

	l, err := elemoLicense.NewLicense(string(data), keys.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse license: %w", err)
	}

	logger.Info(ctx, "license parsed", log.WithValue(l.ID.String()))
	return l, nil
}

func wire(ctx context.Context, cfg *config.Config, logger log.Logger) (*deps, error) {
	tracer := tracing.NoopTracer()
	d := &deps{cfg: cfg, logger: logger, tracer: tracer}

	lic, err := parseLicense(ctx, logger, &cfg.License)
	if err != nil {
		return nil, err
	}

	graphDB, err := initGraphDatabase(ctx, cfg, logger, tracer)
	if err != nil {
		return nil, err
	}
	d.graphDB = graphDB

	relDB, err := initRelationalDatabase(ctx, cfg, logger, tracer)
	if err != nil {
		return nil, err
	}
	d.relDB = relDB

	cacheDB, err := initCacheDatabase(ctx, cfg, logger, tracer)
	if err != nil {
		return nil, err
	}
	d.cacheDB = cacheDB

	searchRepo, err := initSearchRepository(ctx, cfg, logger, tracer)
	if err != nil {
		return nil, err
	}
	d.searchRepo = searchRepo

	s3Client, err := repository.NewS3Client(ctx, &cfg.S3Storage)
	if err != nil {
		return nil, fmt.Errorf("s3 client: %w", err)
	}
	s3Storage, err := repository.NewStorage(
		repository.WithStorageClient(s3Client),
		repository.WithStorageBucket(cfg.S3Storage.Bucket),
		repository.WithStorageLogger(logger.Named("static_file_storage")),
		repository.WithStorageTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("s3 storage: %w", err)
	}
	staticFileRepo, err := repository.NewStaticFileRepository(
		repository.WithS3Storage(s3Storage),
		repository.WithS3RepositoryLogger(logger.Named("static_file_repository")),
		repository.WithS3RepositoryTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("static file repository: %w", err)
	}

	neo4jOpts := func(name string) []repository.Neo4jRepositoryOption {
		return []repository.Neo4jRepositoryOption{
			repository.WithNeo4jDatabase(graphDB),
			repository.WithNeo4jRepositoryLogger(logger.Named(name)),
			repository.WithNeo4jRepositoryTracer(tracer),
		}
	}

	permissionRepo, err := repository.NewNeo4jPermissionRepository(neo4jOpts("permission_repository")...)
	if err != nil {
		return nil, fmt.Errorf("permission repository: %w", err)
	}
	licenseRepo, err := repository.NewNeo4jLicenseRepository(neo4jOpts("license_repository")...)
	if err != nil {
		return nil, fmt.Errorf("license repository: %w", err)
	}
	organizationRepo, err := repository.NewNeo4jOrganizationRepository(neo4jOpts("organization_repository")...)
	if err != nil {
		return nil, fmt.Errorf("organization repository: %w", err)
	}
	roleRepo, err := repository.NewNeo4jRoleRepository(neo4jOpts("role_repository")...)
	if err != nil {
		return nil, fmt.Errorf("role repository: %w", err)
	}
	teamRepo, err := repository.NewNeo4jTeamRepository(neo4jOpts("team_repository")...)
	if err != nil {
		return nil, fmt.Errorf("team repository: %w", err)
	}
	userRepo, err := repository.NewNeo4jUserRepository(neo4jOpts("user_repository")...)
	if err != nil {
		return nil, fmt.Errorf("user repository: %w", err)
	}
	namespaceRepo, err := repository.NewNeo4jNamespaceRepository(neo4jOpts("namespace_repository")...)
	if err != nil {
		return nil, fmt.Errorf("namespace repository: %w", err)
	}
	projectRepo, err := repository.NewNeo4jProjectRepository(neo4jOpts("project_repository")...)
	if err != nil {
		return nil, fmt.Errorf("project repository: %w", err)
	}
	issueRepo, err := repository.NewNeo4jIssueRepository(neo4jOpts("issue_repository")...)
	if err != nil {
		return nil, fmt.Errorf("issue repository: %w", err)
	}
	assignmentRepo, err := repository.NewNeo4jAssignmentRepository(neo4jOpts("assignment_repository")...)
	if err != nil {
		return nil, fmt.Errorf("assignment repository: %w", err)
	}
	labelRepo, err := repository.NewNeo4jLabelRepository(neo4jOpts("label_repository")...)
	if err != nil {
		return nil, fmt.Errorf("label repository: %w", err)
	}
	documentRepo, err := repository.NewNeo4jDocumentRepository(neo4jOpts("document_repository")...)
	if err != nil {
		return nil, fmt.Errorf("document repository: %w", err)
	}
	userTokenRepo, err := repository.NewUserTokenRepository(
		repository.WithPGDatabase(relDB),
		repository.WithPGRepositoryLogger(logger.Named("user_token_repository")),
		repository.WithPGRepositoryTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("user token repository: %w", err)
	}

	permissionService, err := service.NewPermissionService(
		permissionRepo,
		roleRepo,
		service.WithLogger(logger.Named("permission_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("permission service: %w", err)
	}
	d.permissions = permissionService

	licenseService, err := service.NewLicenseService(
		lic,
		licenseRepo,
		permissionService,
		service.WithLogger(logger.Named("license_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("license service: %w", err)
	}

	noopSearch := noopSearchService{}
	realSearch, err := service.NewSearchService(
		searchRepo,
		permissionService,
		nil,
		service.WithLogger(logger.Named("search_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("search service: %w", err)
	}
	d.search = realSearch

	staticFileService, err := service.NewStaticFileService(
		staticFileRepo,
		licenseService,
		service.WithLogger(logger.Named("static_file_service")),
		service.WithTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("static file service: %w", err)
	}

	notificationService := discardNotificationService{}

	namedOpts := func(name string) []service.Option {
		return []service.Option{
			service.WithLogger(logger.Named(name)),
			service.WithTracer(tracer),
		}
	}

	d.users, err = service.NewUserService(
		userRepo,
		userTokenRepo,
		licenseService,
		namedOpts("user_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}
	d.orgs, err = service.NewOrganizationService(
		organizationRepo,
		userRepo,
		userTokenRepo,
		roleRepo,
		permissionService,
		licenseService,
		discardEmailService{},
		notificationService,
		noopSearch,
		namedOpts("organization_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("organization service: %w", err)
	}
	d.teams, err = service.NewTeamService(
		teamRepo,
		permissionService,
		licenseService,
		namedOpts("team_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}
	d.roles, err = service.NewRoleService(
		roleRepo,
		permissionService,
		licenseService,
		organizationRepo,
		notificationService,
		namedOpts("role_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("role service: %w", err)
	}
	d.namespaces, err = service.NewNamespaceService(
		namespaceRepo,
		permissionService,
		licenseService,
		noopSearch,
		namedOpts("namespace_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("namespace service: %w", err)
	}
	d.projects, err = service.NewProjectService(
		projectRepo,
		permissionService,
		licenseService,
		noopSearch,
		namedOpts("project_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("project service: %w", err)
	}
	d.issues, err = service.NewIssueService(
		issueRepo,
		assignmentRepo,
		labelRepo,
		permissionService,
		licenseService,
		noopSearch,
		namedOpts("issue_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("issue service: %w", err)
	}
	d.documents, err = service.NewDocumentService(
		documentRepo,
		licenseService,
		permissionService,
		staticFileService,
		noopSearch,
		namedOpts("document_service")...,
	)
	if err != nil {
		return nil, fmt.Errorf("document service: %w", err)
	}

	return d, nil
}

func initGraphDatabase(
	ctx context.Context,
	cfg *config.Config,
	logger log.Logger,
	tracer tracing.Tracer,
) (*repository.Neo4jDatabase, error) {
	driver, err := repository.NewNeo4jDriver(&cfg.GraphDatabase)
	if err != nil {
		return nil, fmt.Errorf("neo4j driver: %w", err)
	}
	db, err := repository.NewNeo4jDatabase(
		repository.WithNeo4jDriver(driver),
		repository.WithNeo4jDatabaseName(cfg.GraphDatabase.Database),
		repository.WithNeo4jDatabaseLogger(logger.Named("neo4j")),
		repository.WithNeo4jDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("neo4j database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("neo4j ping: %w", err)
	}
	return db, nil
}

func initRelationalDatabase(
	ctx context.Context,
	cfg *config.Config,
	logger log.Logger,
	tracer tracing.Tracer,
) (*repository.PGDatabase, error) {
	pool, err := repository.NewPool(ctx, &cfg.RelationalDatabase)
	if err != nil {
		return nil, fmt.Errorf("postgres pool: %w", err)
	}
	db, err := repository.NewPGDatabase(
		repository.WithDatabasePool(pool),
		repository.WithPGDatabaseLogger(logger.Named("postgres")),
		repository.WithPGDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("postgres database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return db, nil
}

func initCacheDatabase(
	ctx context.Context,
	cfg *config.Config,
	logger log.Logger,
	tracer tracing.Tracer,
) (*repository.RedisDatabase, error) {
	client, err := repository.NewRedisClient(&cfg.CacheDatabase)
	if err != nil {
		return nil, fmt.Errorf("redis client: %w", err)
	}
	db, err := repository.NewRedisDatabase(
		repository.WithRedisClient(client),
		repository.WithRedisDatabaseLogger(logger.Named("redis")),
		repository.WithRedisDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("redis database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return db, nil
}

func initSearchRepository(
	ctx context.Context,
	cfg *config.Config,
	logger log.Logger,
	tracer tracing.Tracer,
) (*repository.MeilisearchSearchRepository, error) {
	client, err := repository.NewMeilisearchClient(&cfg.Search)
	if err != nil {
		return nil, fmt.Errorf("meilisearch client: %w", err)
	}
	db, err := repository.NewSearchDatabase(
		repository.WithSearchClient(client),
		repository.WithSearchDatabaseLogger(logger.Named("meilisearch")),
		repository.WithSearchDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("search database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("meilisearch ping: %w", err)
	}
	repo, err := repository.NewMeilisearchSearchRepository(
		repository.WithSearchDatabase(db),
		repository.WithSearchIndex(cfg.Search.Bucket),
		repository.WithSearchRepositoryLogger(logger.Named("search_repository")),
		repository.WithSearchRepositoryTracer(tracer),
	)
	if err != nil {
		return nil, fmt.Errorf("search repository: %w", err)
	}
	if err := repo.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure search index: %w", err)
	}
	return repo, nil
}

func reindexSearch(ctx context.Context, d *deps) error {
	if err := d.search.Reindex(ctx, service.SearchReindexSources{
		DB: d.graphDB,
	}, service.SearchReindexOptions{
		BatchSize:   d.cfg.Search.ReindexBatchSize,
		Concurrency: d.cfg.Search.ReindexConcurrency,
		DeleteAll:   true,
	}); err != nil {
		return fmt.Errorf("reindex search: %w", err)
	}
	return nil
}
