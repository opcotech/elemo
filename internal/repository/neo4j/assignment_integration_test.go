package neo4j_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type AssignmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *model.User
	testOrg    *model.Organization
	testDoc    *model.Document
	assignment *model.Assignment
}

func (s *AssignmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *AssignmentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.assignment = testModel.NewAssignment(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAssignment), s.assignment.ID)
	s.Assert().NotNil(s.assignment.CreatedAt)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))

	assignment, err := s.AssignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.assignment.ID, assignment.ID)
	s.Assert().Equal(s.assignment.User, assignment.User)
	s.Assert().Equal(s.assignment.Resource, assignment.Resource)
	s.Assert().Equal(s.assignment.Kind, assignment.Kind)
	s.Assert().WithinDuration(*s.assignment.CreatedAt, *assignment.CreatedAt, 100*time.Millisecond)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestGetByUser() {
	assignee := testModel.NewAssignment(s.testUser.ID, s.testDoc.ID, model.AssignmentKindAssignee)
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), assignee))

	reviewer := testModel.NewAssignment(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer)
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), reviewer))

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
	assignee := testModel.NewAssignment(s.testUser.ID, s.testDoc.ID, model.AssignmentKindAssignee)
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), assignee))

	reviewer := testModel.NewAssignment(s.testUser.ID, s.testDoc.ID, model.AssignmentKindReviewer)
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), reviewer))

	assignments, err := s.AssignmentRepo.GetByResource(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 2)

	assignments, err = s.AssignmentRepo.GetByResource(context.Background(), s.testDoc.ID, 0, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 1)

	assignments, err = s.AssignmentRepo.GetByResource(context.Background(), s.testDoc.ID, 1, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 1)

	assignments, err = s.AssignmentRepo.GetByResource(context.Background(), s.testDoc.ID, 2, 1)
	s.Require().NoError(err)
	s.Assert().Len(assignments, 0)
}

func (s *AssignmentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.AssignmentRepo.Create(context.Background(), s.assignment))

	s.Require().NoError(s.AssignmentRepo.Delete(context.Background(), s.assignment.ID))

	_, err := s.AssignmentRepo.Get(context.Background(), s.assignment.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestAssignmentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(AssignmentRepositoryIntegrationTestSuite))
}
