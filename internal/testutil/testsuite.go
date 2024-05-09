package testutil

import (
	"context"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"
	"github.com/opcotech/elemo/internal/repository/redis"
	testContainer "github.com/opcotech/elemo/internal/testutil/container"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

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
	Neo4jDB *neo4j.Database

	AssignmentRepo   *neo4j.AssignmentRepository
	AttachmentRepo   *neo4j.AttachmentRepository
	CommentRepo      *neo4j.CommentRepository
	DocumentRepo     *neo4j.DocumentRepository
	IssueRepo        *neo4j.IssueRepository
	LabelRepo        *neo4j.LabelRepository
	LicenseRepo      *neo4j.LicenseRepository
	NamespaceRepo    *neo4j.NamespaceRepository
	OrganizationRepo *neo4j.OrganizationRepository
	PermissionRepo   *neo4j.PermissionRepository
	ProjectRepo      *neo4j.ProjectRepository
	RoleRepo         *neo4j.RoleRepository
	TodoRepo         *neo4j.TodoRepository
	UserRepo         *neo4j.UserRepository
}

func (s *Neo4jContainerIntegrationTestSuite) BootstrapNeo4jDatabase(ts *ContainerIntegrationTestSuite) {
	testRepo.BootstrapNeo4jDatabase(context.Background(), ts.T(), s.Neo4jDB)
}

func (s *Neo4jContainerIntegrationTestSuite) SetupNeo4j(ts *ContainerIntegrationTestSuite, name string) {
	var err error

	neo4jC, neo4jDBConf := testContainer.NewNeo4jContainer(context.Background(), ts.T(), name)
	ts.AddContainer(neo4jC)

	s.Neo4jDB, _ = testRepo.NewNeo4jDatabase(ts.T(), neo4jDBConf)

	s.AssignmentRepo, err = neo4j.NewAssignmentRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.AttachmentRepo, err = neo4j.NewAttachmentRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.CommentRepo, err = neo4j.NewCommentRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.DocumentRepo, err = neo4j.NewDocumentRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.IssueRepo, err = neo4j.NewIssueRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.LabelRepo, err = neo4j.NewLabelRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.LicenseRepo, err = neo4j.NewLicenseRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.NamespaceRepo, err = neo4j.NewNamespaceRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.OrganizationRepo, err = neo4j.NewOrganizationRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.PermissionRepo, err = neo4j.NewPermissionRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.ProjectRepo, err = neo4j.NewProjectRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.RoleRepo, err = neo4j.NewRoleRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.TodoRepo, err = neo4j.NewTodoRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.UserRepo, err = neo4j.NewUserRepository(neo4j.WithDatabase(s.Neo4jDB))
	ts.Require().NoError(err)

	s.BootstrapNeo4jDatabase(ts)
}

func (s *Neo4jContainerIntegrationTestSuite) CleanupNeo4j(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupNeo4jStore(context.Background(), ts.T(), s.Neo4jDB)
}

// PgContainerIntegrationTestSuite is a test suite which sets up a Postgres
// container to run tests.
type PgContainerIntegrationTestSuite struct {
	PostgresDB *pg.Database
}

func (s *PgContainerIntegrationTestSuite) BootstrapPgDatabase(ts *ContainerIntegrationTestSuite) {
	testRepo.BootstrapPgDatabase(context.Background(), ts.T(), s.PostgresDB)
}

func (s *PgContainerIntegrationTestSuite) SetupPg(ts *ContainerIntegrationTestSuite, name string) {
	pgC, pgDBConf := testContainer.NewPgContainer(context.Background(), ts.T(), name)
	ts.AddContainer(pgC)

	s.PostgresDB, _ = testRepo.NewPgDatabase(ts.T(), pgDBConf)

	s.BootstrapPgDatabase(ts)
}

func (s *PgContainerIntegrationTestSuite) CleanupPg(ts *ContainerIntegrationTestSuite) {
	testRepo.CleanupPgStore(context.Background(), ts.T(), s.PostgresDB)
}

// RedisContainerIntegrationTestSuite is a test suite which sets up a Redis
// container to run tests.
type RedisContainerIntegrationTestSuite struct {
	RedisDB   *redis.Database
	RedisConf *config.CacheDatabaseConfig

	CachedTodoRepo *redis.CachedTodoRepository
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
