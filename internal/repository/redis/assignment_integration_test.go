package redis_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/redis"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type CachedAssignmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *model.User
	testOrg        *model.Organization
	testIssue      *model.Issue
	assignment     *model.Assignment
	assignmentRepo *redis.CachedAssignmentRepository
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.assignmentRepo, _ = redis.NewCachedAssignmentRepository(s.AssignmentRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testIssue = testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testUser.ID, s.testIssue))

	s.assignment = testModel.NewAssignment(s.testUser.ID, s.testIssue.ID, model.AssignmentKindReviewer)

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.assignmentRepo.Create(context.Background(), s.assignment))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAssignment), s.assignment.ID)
	s.Assert().NotNil(s.assignment.CreatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))

	original, err := s.AssignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Require().NoError(err)

	usingCache, err := s.assignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.assignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGetByResource() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), testModel.NewAssignment(s.testUser.ID, s.testIssue.ID, model.AssignmentKindReviewer)))

	originalAssignments, err := s.AssignmentRepo.GetByResource(context.Background(), s.testIssue.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheAssignments, err := s.assignmentRepo.GetByResource(context.Background(), s.testIssue.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalAssignments, usingCacheAssignments)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedAssignments, err := s.assignmentRepo.GetByResource(context.Background(), s.testIssue.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheAssignments), len(cachedAssignments))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGetByUser() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), testModel.NewAssignment(s.testUser.ID, s.testIssue.ID, model.AssignmentKindReviewer)))

	originalAssignments, err := s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheAssignments, err := s.assignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalAssignments, usingCacheAssignments)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedAssignments, err := s.assignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheAssignments), len(cachedAssignments))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))

	_, err := s.assignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.assignmentRepo.Delete(context.Background(), s.assignment.ID))

	_, err = s.assignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedAssignmentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CachedAssignmentRepositoryIntegrationTestSuite))
}
