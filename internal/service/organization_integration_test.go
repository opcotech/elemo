package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type OrganizationServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	organizationService service.OrganizationService

	owner        *model.User
	organization *model.Organization

	ctx context.Context
}

func (s *OrganizationServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		service.WithUserRepository(s.UserRepo),
		service.WithOrganizationRepository(s.OrganizationRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *OrganizationServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *OrganizationServiceIntegrationTestSuite) TestCreate() {
	err := s.organizationService.Create(s.ctx, s.owner.ID, s.organization)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.organization.ID)
	s.Assert().NotNil(s.organization.CreatedAt)
	s.Assert().Nil(s.organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(s.organization.ID, organization.ID)
	s.Assert().Equal(s.organization.Name, organization.Name)
	s.Assert().Equal(s.organization.Logo, organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().ElementsMatch(s.organization.Namespaces, organization.Namespaces)
	s.Assert().ElementsMatch(s.organization.Teams, organization.Teams)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
	s.Assert().Equal(s.organization.CreatedAt, organization.CreatedAt)
	s.Assert().Equal(s.organization.UpdatedAt, organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))

	organizations, err := s.organizationService.GetAll(s.ctx, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 3)

	organizations, err = s.organizationService.GetAll(s.ctx, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 2)

	organizations, err = s.organizationService.GetAll(s.ctx, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 2)

	organizations, err = s.organizationService.GetAll(s.ctx, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 1)

	organizations, err = s.organizationService.GetAll(s.ctx, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 0)
}

func (s *OrganizationServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	patch := map[string]any{
		"name": "new name",
		"logo": "https://example.com/static/new-logo.png",
	}

	organization, err := s.organizationService.Update(s.ctx, s.organization.ID, patch)
	s.Require().NoError(err)
	s.Assert().Equal(patch["name"], organization.Name)
	s.Assert().Equal(patch["logo"], organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().ElementsMatch(s.organization.Namespaces, organization.Namespaces)
	s.Assert().ElementsMatch(s.organization.Teams, organization.Teams)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
	s.Assert().Equal(s.organization.CreatedAt, organization.CreatedAt)
	s.Assert().NotNil(organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	err = s.organizationService.AddMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, member.ID}, organization.Members)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGetMembers() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	members, err := s.organizationService.GetMembers(s.ctx, s.organization.ID)
	s.Require().NoError(err)

	memberIDs := make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, memberIDs)
	s.Assert().Len(members, 1)

	// Owner should have roles (includes virtual roles based on permissions)
	s.Assert().NotNil(members[0].Roles)
	s.Assert().NotEmpty(members[0].Roles)

	// Owner should have "owner" role (virtual role based on permissions)
	s.Assert().Contains(members[0].Roles, "Owner")
}

func (s *OrganizationServiceIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	err := s.organizationService.AddMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, member.ID}, organization.Members)

	err = s.organizationService.RemoveMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
}

func (s *OrganizationServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	err := s.organizationService.Delete(s.ctx, s.organization.ID, false)
	s.Require().NoError(err)

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(model.OrganizationStatusDeleted, organization.Status)

	err = s.organizationService.Delete(s.ctx, s.organization.ID, true)
	s.Require().NoError(err)

	_, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func TestOrganizationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceIntegrationTestSuite))
}
