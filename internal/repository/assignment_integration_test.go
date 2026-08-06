package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type AssignmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	testDoc    *repository.Document
	createOpts repository.CreateAssignmentOpts
}

func (s *AssignmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *AssignmentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestCreate() {
	assignment, err := s.AssignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAssignment), assignment.ID)
	s.Assert().NotNil(assignment.CreatedAt)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.AssignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	assignment, err := s.AssignmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, assignment.ID)
	s.Assert().Equal(s.createOpts.User, assignment.User)
	s.Assert().Equal(s.createOpts.Resource, assignment.Resource)
	s.Assert().Equal(s.createOpts.Kind, assignment.Kind)
	s.Assert().WithinDuration(*created.CreatedAt, *assignment.CreatedAt, 100*time.Millisecond)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestGetByUser() {
	_, err := s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testDoc.ID, model.AssignmentKindAssignee))
	s.Require().NoError(err)
	_, err = s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer))
	s.Require().NoError(err)

	assignments, err := s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 2)

	assignments, err = s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 1)

	assignments, err = s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 1, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 1)

	assignments, err = s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 2, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 0)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestGetByResource() {
	_, err := s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testDoc.ID, model.AssignmentKindAssignee))
	s.Require().NoError(err)
	_, err = s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer))
	s.Require().NoError(err)

	assignments, err := s.AssignmentRepo.GetByResource(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 2)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.AssignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.AssignmentRepo.Delete(context.Background(), created.ID))

	_, err = s.AssignmentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestAssignmentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentRepositoryIntegrationTestSuite))
}

type CachedAssignmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *repository.User
	testOrg        *repository.Organization
	testNamespace  *repository.Namespace
	testProject    *repository.Project
	testIssue      *repository.Issue
	createOpts     repository.CreateAssignmentOpts
	assignmentRepo *repository.RedisCachedAssignmentRepository
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.assignmentRepo, _ = repository.NewCachedAssignmentRepository(s.AssignmentRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testNamespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)

	s.testProject, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID))
	s.Require().NoError(err)

	s.testIssue, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testIssue.ID, model.AssignmentKindReviewer)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestCreate() {
	assignment, err := s.assignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAssignment), assignment.ID)
	s.Assert().NotNil(assignment.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.assignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.AssignmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	usingCache, err := s.assignmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGetByUser() {
	_, err := s.assignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(s.testUser.ID, s.testIssue.ID, model.AssignmentKindAssignee))
	s.Require().NoError(err)
	_, err = s.assignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.AssignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	usingCache, err := s.assignmentRepo.GetByUser(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestGetByResource() {
	_, err := s.assignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.AssignmentRepo.GetByResource(context.Background(), s.testIssue.ID, 0, 10)
	s.Require().NoError(err)

	usingCache, err := s.assignmentRepo.GetByResource(context.Background(), s.testIssue.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAssignmentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.assignmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	_, err = s.assignmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.assignmentRepo.Delete(context.Background(), created.ID))

	_, err = s.assignmentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedAssignmentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedAssignmentRepositoryIntegrationTestSuite))
}
