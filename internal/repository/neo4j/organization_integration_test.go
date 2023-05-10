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

type OrganizationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser     *model.User
	organization *model.Organization
}

func (s *OrganizationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *OrganizationRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.organization = testModel.NewOrganization()
}

func (s *OrganizationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeOrganization), s.organization.ID)
	s.Assert().NotNil(s.organization.CreatedAt)
	s.Assert().Nil(s.organization.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	organization, err := s.OrganizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.organization.ID, organization.ID)
	s.Assert().Equal(s.organization.Name, organization.Name)
	s.Assert().Equal(s.organization.Email, organization.Email)
	s.Assert().Equal(s.organization.Logo, organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().Equal(s.organization.Namespaces, organization.Namespaces)
	s.Assert().Equal(s.organization.Teams, organization.Teams)
	s.Assert().Equal([]model.ID{s.testUser.ID}, organization.Members)
	s.Assert().WithinDuration(*s.organization.CreatedAt, *organization.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.organization.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, testModel.NewOrganization()))
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, testModel.NewOrganization()))

	organizations, err := s.OrganizationRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Require().Len(organizations, 3)

	organizations, err = s.OrganizationRepo.GetAll(context.Background(), 1, 2)
	s.Require().NoError(err)
	s.Require().Len(organizations, 2)

	organizations, err = s.OrganizationRepo.GetAll(context.Background(), 2, 2)
	s.Require().NoError(err)
	s.Require().Len(organizations, 1)

	organizations, err = s.OrganizationRepo.GetAll(context.Background(), 3, 2)
	s.Require().NoError(err)
	s.Require().Len(organizations, 0)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	patch := map[string]any{
		"name":  "new name",
		"email": testutil.GenerateEmail(10),
	}

	organization, err := s.OrganizationRepo.Update(context.Background(), s.organization.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.organization.ID, organization.ID)
	s.Assert().Equal(patch["name"], organization.Name)
	s.Assert().Equal(patch["email"], organization.Email)
	s.Assert().Equal(s.organization.Logo, organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().Equal(s.organization.Namespaces, organization.Namespaces)
	s.Assert().Equal(s.organization.Teams, organization.Teams)
	s.Assert().Equal([]model.ID{s.testUser.ID}, organization.Members)
	s.Assert().WithinDuration(*s.organization.CreatedAt, *organization.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(organization.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))

	organization, err := s.OrganizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().ElementsMatch([]model.ID{s.testUser.ID, member.ID}, organization.Members)
	s.Assert().Nil(organization.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))
	s.Require().NoError(s.OrganizationRepo.RemoveMember(context.Background(), s.organization.ID, s.testUser.ID))

	organization, err := s.OrganizationRepo.Get(context.Background(), s.organization.ID)
	s.Require().NoError(err)

	s.Assert().ElementsMatch([]model.ID{member.ID}, organization.Members)
	s.Assert().Nil(organization.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.organization))

	s.Require().NoError(s.OrganizationRepo.Delete(context.Background(), s.organization.ID))

	_, err := s.OrganizationRepo.Get(context.Background(), s.organization.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestOrganizationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationRepositoryIntegrationTestSuite))
}
