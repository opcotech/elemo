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

type RoleServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	roleService         service.RoleService
	organizationService service.OrganizationService

	owner        *model.User
	role         *model.Role
	organization *model.Organization

	ctx context.Context
}

func (s *RoleServiceIntegrationTestSuite) SetupSuite() {
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

	s.roleService, err = service.NewRoleService(
		service.WithRoleRepository(s.RoleRepo),
		service.WithUserRepository(s.UserRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
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

func (s *RoleServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	s.role = testModel.NewRole()
}

func (s *RoleServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *RoleServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *RoleServiceIntegrationTestSuite) TestCreate() {
	err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.role.ID)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(s.role.ID, role.ID)
	s.Assert().Equal(s.role.Name, role.Name)
	s.Assert().Equal(s.role.Description, role.Description)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, role.Members)
	s.Assert().ElementsMatch(s.role.Permissions, role.Permissions)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	roles, err := s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(roles, 3)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 1)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 0)
}

func (s *RoleServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	updated, err := s.roleService.Update(s.ctx, s.role.ID, s.organization.ID, patch)
	s.Require().NoError(err)
	s.Assert().Equal(s.role.ID, updated.ID)
	s.Assert().Equal(patch["name"], updated.Name)
	s.Assert().Equal(patch["description"], updated.Description)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, updated.Members)
	s.Assert().ElementsMatch(s.role.Permissions, updated.Permissions)
	s.Assert().NotNil(updated.CreatedAt)
	s.Assert().NotNil(updated.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGetMembers() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	members, err := s.roleService.GetMembers(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)

	memberIDs := make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, memberIDs)

	err = s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	members, err = s.roleService.GetMembers(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	memberIDs = make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, user.ID}, memberIDs)
}

func (s *RoleServiceIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	err := s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, user.ID}, role.Members)
}

func (s *RoleServiceIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	err := s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	err = s.roleService.RemoveMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, role.Members)
}

func (s *RoleServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	err := s.roleService.Delete(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)

	_, err = s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestRoleServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceIntegrationTestSuite))
}
