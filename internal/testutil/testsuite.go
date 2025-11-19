package testutil

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository"
	testConfig "github.com/opcotech/elemo/internal/testutil/config"
	testContainer "github.com/opcotech/elemo/internal/testutil/container"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type ConfigurationTestSuite struct {
	Config *config.Config
}

func (s *ConfigurationTestSuite) LoadConfig() {
	s.Config = testConfig.Conf
}

// ContainerIntegrationTestSuite is a test suite which uses a container to run
// tests.
type ContainerIntegrationTestSuite struct {
	suite.Suite

	// containers is a list of containers to be cleaned up after the test.
	containers []testcontainers.Container
}

// AddContainer adds a container to the list of containers to be cleaned up
// after the test.
func (s *ContainerIntegrationTestSuite) AddContainer(container testcontainers.Container) {
	s.containers = append(s.containers, container)
}

// CleanupContainers cleans up any containers created during the test.
func (s *ContainerIntegrationTestSuite) CleanupContainers() {
	for i, c := range s.containers {
		s.containers = s.containers[:i]
		if err := c.Terminate(context.Background()); err != nil {
			s.T().Errorf("failed to terminate container: %s", err.Error())
		}
	}
}

// Neo4jContainerIntegrationTestSuite is a test suite which sets up a Neo4j
// container to run tests.
type Neo4jContainerIntegrationTestSuite struct {
	Neo4jDB *repository.Neo4jDatabase

	AssignmentRepo   *repository.Neo4jAssignmentRepository
	AttachmentRepo   *repository.Neo4jAttachmentRepository
	CommentRepo      *repository.Neo4jCommentRepository
	DocumentRepo     *repository.Neo4jDocumentRepository
	IssueRepo        *repository.Neo4jIssueRepository
	LabelRepo        *repository.Neo4jLabelRepository
	LicenseRepo      *repository.Neo4jLicenseRepository
	NamespaceRepo    *repository.Neo4jNamespaceRepository
	OrganizationRepo *repository.Neo4jOrganizationRepository
	PermissionRepo   *repository.Neo4jPermissionRepository
	ProjectRepo      *repository.Neo4jProjectRepository
	RoleRepo         *repository.Neo4jRoleRepository
	TodoRepo         *repository.Neo4jTodoRepository
	UserRepo         *repository.Neo4jUserRepository
}

func (s *Neo4jContainerIntegrationTestSuite) BootstrapNeo4jDatabase(ts *ContainerIntegrationTestSuite) {
	testRepo.BootstrapNeo4jDatabase(context.Background(), ts.T(), s.Neo4jDB)
}

func (s *Neo4jContainerIntegrationTestSuite) SetupNeo4j(ts *ContainerIntegrationTestSuite, name string) {
	var err error

	neo4jC, neo4jDBConf := testContainer.NewNeo4jContainer(context.Background(), ts.T(), name)
	ts.AddContainer(neo4jC)

	s.Neo4jDB, _ = testRepo.NewNeo4jDatabase(ts.T(), neo4jDBConf)

	s.AssignmentRepo, err = repository.NewNeo4jAssignmentRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.AttachmentRepo, err = repository.NewNeo4jAttachmentRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.CommentRepo, err = repository.NewNeo4jCommentRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.DocumentRepo, err = repository.NewNeo4jDocumentRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.IssueRepo, err = repository.NewNeo4jIssueRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.LabelRepo, err = repository.NewNeo4jLabelRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.LicenseRepo, err = repository.NewNeo4jLicenseRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.NamespaceRepo, err = repository.NewNeo4jNamespaceRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.OrganizationRepo, err = repository.NewNeo4jOrganizationRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.PermissionRepo, err = repository.NewNeo4jPermissionRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.ProjectRepo, err = repository.NewNeo4jProjectRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.RoleRepo, err = repository.NewNeo4jRoleRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.TodoRepo, err = repository.NewNeo4jTodoRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.UserRepo, err = repository.NewNeo4jUserRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.BootstrapNeo4jDatabase(ts)
}

func (s *Neo4jContainerIntegrationTestSuite) CleanupNeo4j(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupNeo4jStore(context.Background(), ts.T(), s.Neo4jDB)
}

// PgContainerIntegrationTestSuite is a test suite which sets up a Postgres
// container to run tests.
type PgContainerIntegrationTestSuite struct {
	PostgresDB *repository.PGDatabase

	NotificationRepo    *repository.PGNotificationRepository
	UserTokenRepository *repository.PGUserTokenRepository
}

func (s *PgContainerIntegrationTestSuite) BootstrapPgDatabase(ts *ContainerIntegrationTestSuite) {
	testRepo.BootstrapPgDatabase(context.Background(), ts.T(), s.PostgresDB)
}

func (s *PgContainerIntegrationTestSuite) SetupPg(ts *ContainerIntegrationTestSuite, name string) {
	var err error

	pgC, pgDBConf := testContainer.NewPgContainer(context.Background(), ts.T(), name)
	ts.AddContainer(pgC)

	s.PostgresDB, _ = testRepo.NewPgDatabase(ts.T(), pgDBConf)

	s.NotificationRepo, err = repository.NewNotificationRepository(repository.WithPGDatabase(s.PostgresDB))
	ts.Require().NoError(err)

	s.UserTokenRepository, err = repository.NewUserTokenRepository(repository.WithPGDatabase(s.PostgresDB))
	ts.Require().NoError(err)

	s.BootstrapPgDatabase(ts)
}

func (s *PgContainerIntegrationTestSuite) CleanupPg(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupPgStore(context.Background(), ts.T(), s.PostgresDB)
}

// RedisContainerIntegrationTestSuite is a test suite which sets up a Redis
// container to run tests.
type RedisContainerIntegrationTestSuite struct {
	RedisDB   *repository.RedisDatabase
	RedisConf *config.CacheDatabaseConfig

	CachedTodoRepo *repository.RedisCachedTodoRepository
}

func (s *RedisContainerIntegrationTestSuite) SetupRedis(ts *ContainerIntegrationTestSuite, name string) {
	var redisC testcontainers.Container
	redisC, s.RedisConf = testContainer.NewRedisContainer(context.Background(), ts.T(), name)
	ts.AddContainer(redisC)

	s.RedisDB, _ = testRepo.NewRedisDatabase(ts.T(), s.RedisConf)
}

func (s *RedisContainerIntegrationTestSuite) CleanupRedis(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupRedisStore(context.Background(), ts.T(), s.RedisDB)
}

func (s *RedisContainerIntegrationTestSuite) GetKeys(ts *ContainerIntegrationTestSuite, pattern string) []string {
	keys, err := s.RedisDB.GetClient().Keys(context.Background(), pattern).Result()
	ts.Require().NoError(err)
	return keys
}

// LocalStackContainerIntegrationTestSuite is a test suite which sets up a
// LocalStack container to run tests.
type LocalStackContainerIntegrationTestSuite struct {
	S3Storage *repository.S3Storage

	StaticFileRepository repository.StaticFileRepository
}

func (s *LocalStackContainerIntegrationTestSuite) BootstrapLocalStack(ts *ContainerIntegrationTestSuite) {
	testRepo.BootstrapS3Storage(context.Background(), ts.T(), s.S3Storage)
}

func (s *LocalStackContainerIntegrationTestSuite) SetupLocalStack(ts *ContainerIntegrationTestSuite, name string) {
	var err error

	localStackC, localStackConf := testContainer.NewLocalStackContainer(context.Background(), ts.T(), name)
	ts.AddContainer(localStackC)

	s.S3Storage = testRepo.NewS3Storage(ts.T(), localStackConf)

	s.StaticFileRepository, err = repository.NewStaticFileRepository(repository.WithS3Storage(s.S3Storage))
	ts.Require().NoError(err)

	s.BootstrapLocalStack(ts)
}

func (s *LocalStackContainerIntegrationTestSuite) CleanupLocalStack(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupS3Storage(context.Background(), ts.T(), s.S3Storage)
}
